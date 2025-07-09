# SAML Proxy Setup Guide with InfraLib

This guide walks you through setting up a SAML proxy using InfraLib in three main steps:

1. Configure AWS IAM Identity Center Application
2. Create AWS Secrets Manager Object  
3. Deploy SAML Proxy with InfraLib

## Step 1: Configure AWS IAM Identity Center Application

### 1.1 Create SAML 2.0 Application

1. Navigate to your **AWS root account** where IAM Identity Center is configured
2. Go to **IAM Identity Center** → **Applications**
3. Click **Add application** → **I have an application I want to set up** → **SAML 2.0**
4. Configure the application with the following settings:
   * **Application name**: `<app_name>-saml-proxy`
   * **Application ACS URL**: `https://<app_hostname>/mellon/postResponse`
   * **Application SAML audience**: `https://<app_hostname>`

### 1.2 Configure Attribute Mappings

1. In the **Attribute mappings** section, add the following mapping:
   * **Subject**: `${user:email}` → `emailAddress`

> **Note**: Additional attribute mappings may be required depending on your application's needs.

### 1.3 Download Metadata and Assign Users

1. **Download** the IAM Identity Center SAML metadata file (you'll need this in Step 2)
2. **Rename** the downloaded SAML metadata file to `saml_idp.xml`
3. **Assign users or groups** to the newly created application
4. **Save** the application configuration

## Step 2: Create AWS Secrets Manager Object

### 2.1 Download Required Script

Download the metadata creation script from the InfraLib repository:

```bash
curl -O https://raw.githubusercontent.com/entigolabs/entigo-infralib/main/modules/k8s/saml-proxy/create_metadata.sh
chmod +x create_metadata.sh
```

### 2.2 Prepare Environment

1. Ensure both files are in the same directory:
   * `create_metadata.sh` (downloaded above)
   * IAM Identity Center SAML metadata file (downloaded in Step 1.3) (`saml_idp.xml`)
2. Configure AWS credentials for the target account where the SAML proxy will be deployed.

### 2.3 Execute Script

Run the script with the following parameters:

```bash
./create_metadata.sh https://<app_hostname> https://<app_hostname>/mellon <app_name>-saml-proxy <aws_region>
```

**Parameter explanation:**
* `https://<app_hostname>`: Your application's hostname
* `https://<app_hostname>/mellon`: SAML endpoint URL
* `<app_name>-saml-proxy`: Name for the secret (should match your application name)
* `<aws_region>`: Target AWS region

### 2.4 Verify Secret Creation

Confirm the secret was created successfully:

```bash
aws secretsmanager describe-secret --secret-id <app_name>-saml-proxy --region <aws_region>
```

## Step 3: Deploy SAML Proxy with InfraLib

### 3.1 Update InfraLib Configuration

Add the following configuration to your InfraLib configuration file:

```yaml
applications:
  - name: <app_name>-saml-proxy
    source: saml-proxy
    inputs:
      ingress:
        hostName: "<app_hostname>"
      targetService: "http://<target_service>.<target_service_namespace>"
```

**Configuration parameters:**
* `<app_name>`: Your application identifier
* `<app_hostname>`: The hostname where your application will be accessible
* `<target_service>`: The Kubernetes service name you want to protect with SAML
* `<target_service_namespace>`: The namespace where your target service is located

### 3.2 Deploy with InfraLib

Execute the InfraLib agent to deploy the configuration.

### 3.3 Verify Deployment

1. Check that the `<app_name>-saml-proxy` application appears in **ArgoCD**
2. Check that the application is visible in **AWS SSO Portal**
3. Verify the application is in a healthy state
4. Test SAML authentication by accessing `https://<app_hostname>` or clicking the app from AWS SSO Portal.

## Troubleshooting

If you encounter issues:

1. **Verify IAM Identity Center configuration**: Ensure the application is properly configured and users are assigned
2. **Check AWS Secrets Manager**: Confirm the secret exists and contains the correct metadata
3. **Review ArgoCD logs**: Check for any deployment errors in ArgoCD
4. **Validate network connectivity**: Ensure the target service is accessible from the SAML proxy

## Security Considerations

- Ensure proper SSL/TLS certificates are configured for your hostname
- Review and limit user/group access in IAM Identity Center
- Monitor SAML proxy logs for security events
- Keep the SAML metadata up to date if IAM Identity Center configuration changes