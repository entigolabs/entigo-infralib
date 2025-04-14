#!/bin/bash
# Script to convert AWS Config rules from CloudFormation to Terraform

# Check for dependencies
command -v curl >/dev/null 2>&1 || { echo "Error: curl is required but not installed"; exit 1; }
command -v yq >/dev/null 2>&1 || { echo "Error: yq is required. Install with: brew install yq"; exit 1; }

# Get URL from command line or use default
URL=${1:-"https://raw.githubusercontent.com/awslabs/aws-config-rules/master/aws-config-conformance-packs/Operational-Best-Practices-for-CIS-AWS-v1.4-Level1.yaml"}
OUTPUT="rules.tf"

echo "# Converting CloudFormation Config rules to Terraform"
echo "# Source: $URL"
echo ""

# Download the YAML file
echo "Downloading CloudFormation template..."
TEMP_FILE=$(mktemp)
curl -s "$URL" -o "$TEMP_FILE"

# Start output file
echo "# AWS Config Rules converted from CloudFormation" > "$OUTPUT"
echo "# Source: $URL" >> "$OUTPUT"
echo "" >> "$OUTPUT"

# Process each resource in the YAML file
echo "Processing resources..."
for RESOURCE in $(yq '.Resources | keys | .[]' "$TEMP_FILE" 2>/dev/null); do
  # Get resource type
  TYPE=$(yq ".Resources.$RESOURCE.Type" "$TEMP_FILE" 2>/dev/null)
  
  # Only process AWS Config Rules
  if [ "$TYPE" == "AWS::Config::ConfigRule" ]; then
    echo "Converting rule: $RESOURCE"
    
    # Get rule properties
    NAME=$(yq ".Resources.$RESOURCE.Properties.ConfigRuleName // \"$RESOURCE\"" "$TEMP_FILE" 2>/dev/null)
    OWNER=$(yq ".Resources.$RESOURCE.Properties.Source.Owner // \"AWS\"" "$TEMP_FILE" 2>/dev/null)
    SOURCE_ID=$(yq ".Resources.$RESOURCE.Properties.Source.SourceIdentifier" "$TEMP_FILE" 2>/dev/null)
    
    # Convert resource name to valid Terraform resource name (lowercase with underscores)
    TF_NAME=$(echo "$RESOURCE" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/_/g')
    
    # Create the resource
    echo "resource \"aws_config_config_rule\" \"$TF_NAME\" {" >> "$OUTPUT"
    echo "  name  = \"$NAME\"" >> "$OUTPUT"
    echo "" >> "$OUTPUT"
    echo "  source {" >> "$OUTPUT"
    echo "    owner             = \"$OWNER\"" >> "$OUTPUT"
    echo "    source_identifier = \"$SOURCE_ID\"" >> "$OUTPUT"
    echo "  }" >> "$OUTPUT"
    
    # Handle input parameters if they exist
    if yq -e ".Resources.$RESOURCE.Properties.InputParameters" "$TEMP_FILE" &>/dev/null; then
      echo "" >> "$OUTPUT"
      echo "  input_parameters = jsonencode({" >> "$OUTPUT"
      
      PARAM_COUNT=0
      for PARAM in $(yq ".Resources.$RESOURCE.Properties.InputParameters | keys | .[]" "$TEMP_FILE" 2>/dev/null); do
        PARAM_VALUE=$(yq ".Resources.$RESOURCE.Properties.InputParameters.$PARAM" "$TEMP_FILE")
        
        # Check for CloudFormation functions
        if [[ "$PARAM_VALUE" == *"Fn::"* || "$PARAM_VALUE" == *"Ref:"* ]]; then
          # Handle CloudFormation parameters based on common naming patterns
          case "$PARAM" in
            maxAccessKeyAge)
              PARAM_VALUE="90"
              ;;
            MaxPasswordAge)
              PARAM_VALUE="90"
              ;;
            MinimumPasswordLength)
              PARAM_VALUE="14"
              ;;
            PasswordReusePrevention)
              PARAM_VALUE="24"
              ;;
            RequireLowercaseCharacters|RequireNumbers|RequireSymbols|RequireUppercaseCharacters)
              PARAM_VALUE="true"
              ;;
            policyARN)
              PARAM_VALUE="\"arn:aws:iam::aws:policy/AWSSupportAccess\""
              ;;
            maxCredentialUsageAge)
              PARAM_VALUE="45"
              ;;
            blockedPort3)
              PARAM_VALUE="3389"
              ;;
            BlockPublicAcls|BlockPublicPolicy|IgnorePublicAcls|RestrictPublicBuckets)
              PARAM_VALUE="true"
              ;;
            isMfaDeleteEnabled)
              PARAM_VALUE="true"
              ;;
            *)
              PARAM_VALUE="\"REPLACE_ME\""
              ;;
          esac
        elif [[ "$PARAM_VALUE" =~ ^[0-9]+$ ]]; then
          # Numeric values don't need quotes
          :
        elif [[ "$PARAM_VALUE" =~ ^(true|false)$ ]]; then
          # Boolean values don't need quotes
          :
        else
          # String values need quotes
          PARAM_VALUE="\"$PARAM_VALUE\""
        fi
        
        # Add comma for all but the first parameter
        if [ $PARAM_COUNT -gt 0 ]; then
          echo "," >> "$OUTPUT"
        fi
        
        echo -n "    $PARAM = $PARAM_VALUE" >> "$OUTPUT"
        PARAM_COUNT=$((PARAM_COUNT + 1))
      done
      
      echo "" >> "$OUTPUT"
      echo "  })" >> "$OUTPUT"
    fi
    
    # Close the resource
    echo "}" >> "$OUTPUT"
    echo "" >> "$OUTPUT"
  fi
done

echo "Conversion complete! Terraform code written to $OUTPUT"
rm "$TEMP_FILE"