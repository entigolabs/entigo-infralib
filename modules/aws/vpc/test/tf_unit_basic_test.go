package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/retry"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestVpc(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"aws_biz":   testVpcBiz,
		"aws_pri":   testVpcPri,
		"aws_spoke": testVpcSpoke,
	})
}

func testVpcBiz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_id"), "vpc_id was not returned")

	privateSubnets := outputs.StringList(t, "vpc__private_subnets")
	assert.Equal(t, 2, len(privateSubnets), "Wrong number of private_subnets returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__public_subnets")), "Wrong number of public_subnets returned")
	assert.Equal(t, 0, len(outputs.StringList(t, "vpc__intra_subnets")), "Wrong number of intra_subnets returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__database_subnets")), "Wrong number of database_subnets returned")
	assert.NotEmpty(t, outputs.String(t, "vpc__database_subnet_group"), "database_subnet_group was not returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__elasticache_subnets")), "Wrong number of elasticache_subnets returned")
	assert.NotEmpty(t, outputs.String(t, "vpc__elasticache_subnet_group"), "elasticache_subnet_group was not returned")

	privateSubnetCidrs := outputs.StringList(t, "vpc__private_subnets_cidr_blocks")
	assert.Equal(t, []string{"10.146.32.0/21", "10.146.40.0/21"}, privateSubnetCidrs, "Wrong private_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.146.4.0/24", "10.146.5.0/24"}, outputs.StringList(t, "vpc__public_subnets_cidr_blocks"), "Wrong public_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.146.16.0/22", "10.146.20.0/22"}, outputs.StringList(t, "vpc__database_subnets_cidr_blocks"), "Wrong database_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.146.0.0/26", "10.146.0.64/26"}, outputs.StringList(t, "vpc__elasticache_subnets_cidr_blocks"), "Wrong elasticache_subnets_cidr_blocks returned")

	// Without a subnet split mode the control, service and compute subnets are the private ones.
	assert.Equal(t, privateSubnetCidrs, outputs.StringList(t, "vpc__control_subnets_cidr_blocks"), "control_subnets_cidr_blocks must equal private_subnets_cidr_blocks")
	assert.Equal(t, privateSubnetCidrs, outputs.StringList(t, "vpc__service_subnets_cidr_blocks"), "service_subnets_cidr_blocks must equal private_subnets_cidr_blocks")
	assert.Equal(t, privateSubnetCidrs, outputs.StringList(t, "vpc__compute_subnets_cidr_blocks"), "compute_subnets_cidr_blocks must equal private_subnets_cidr_blocks")
	assert.Equal(t, privateSubnets, outputs.StringList(t, "vpc__control_subnets"), "control_subnets must equal private_subnets")
	assert.Equal(t, privateSubnets, outputs.StringList(t, "vpc__service_subnets"), "service_subnets must equal private_subnets")
	assert.Equal(t, privateSubnets, outputs.StringList(t, "vpc__compute_subnets"), "compute_subnets must equal private_subnets")
	assert.Equal(t, 0, len(outputs.StringList(t, "vpc__intra_subnets_cidr_blocks")), "Wrong intra_subnets_cidr_blocks returned")

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_ipv6_cidr_block"), "vpc_ipv6_cidr_block was not returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__public_subnets_ipv6_cidr_blocks")), "Wrong number of public_subnets_ipv6_cidr_blocks returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__private_subnets_ipv6_cidr_blocks")), "Wrong number of private_subnets_ipv6_cidr_blocks returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__database_subnets_ipv6_cidr_blocks")), "Wrong number of database_subnets_ipv6_cidr_blocks returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__elasticache_subnets_ipv6_cidr_blocks")), "Wrong number of elasticache_subnets_ipv6_cidr_blocks returned")
	assert.NotEmpty(t, outputs.StringList(t, "vpc__private_ipv6_egress_route_ids"), "private_ipv6_egress_route_ids was not returned")

	// biz has the kms module, so the created flow log bucket uses the telemetry key
	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_flow_log_id"), "vpc_flow_log_id was not returned")
	assert.Equal(t, "s3", outputs.String(t, "vpc__vpc_flow_log_destination_type"), "Flow logs must go to S3")
	flowLogBucket := outputs.String(t, "vpc__vpc_flow_log_bucket_name")
	assert.NotEmpty(t, flowLogBucket, "Flow log bucket must be created when flow_log_s3_destination_arn is not set")
	assertFlowLogEncryption(t, e, flowLogBucket, outputs.String(t, "vpc__vpc_owner_id"), outputs.String(t, "vpc__vpc_flow_log_id"), "aws:kms")

	endpoints := getEndpoints(t, outputs)
	assert.ElementsMatch(t, []string{"s3", "ecr_api", "ecr_dkr", "ec2", "sts", "efs"}, keys(endpoints), "Wrong vpc_endpoints returned")
	assert.Contains(t, getEndpointPolicy(t, e, endpoints["efs"]), "CustomEfsPolicy", "Custom efs endpoint policy was not applied")
	for key, action := range map[string]string{"s3": "s3:*", "ecr_api": "ecr:*", "ecr_dkr": "ecr:*", "ec2": "ec2:*", "sts": "sts:*"} {
		policy := getEndpointPolicy(t, e, endpoints[key])
		assertServicePolicy(t, policy, key, action)
		assert.Contains(t, policy, "o-cipguslp2x", "%s endpoint policy must trust the organization", key)
	}
	assert.Contains(t, getEndpointPolicy(t, e, endpoints["s3"]), "starport-layer-bucket", "s3 endpoint policy must allow the ECR layer bucket")
	assert.Contains(t, getEndpointPolicy(t, e, endpoints["s3"]), "OwnBucketsRead", "s3 endpoint policy must allow anonymous reads of own buckets")
	assert.Contains(t, getEndpointPolicy(t, e, endpoints["sts"]), "sts:AssumeRoleWithWebIdentity", "sts endpoint policy must allow IRSA")
}

func testVpcPri(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_id"), "vpc_id was not returned")

	privateSubnets := outputs.StringList(t, "vpc__private_subnets")
	assert.Equal(t, 3, len(privateSubnets), "Wrong number of private_subnets returned")
	assert.Equal(t, 3, len(outputs.StringList(t, "vpc__public_subnets")), "Wrong number of public_subnets returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__intra_subnets")), "Wrong number of intra_subnets returned")
	assert.Equal(t, 2, len(outputs.StringList(t, "vpc__database_subnets")), "Wrong number of database_subnets returned")
	assert.Equal(t, 0, len(outputs.StringList(t, "vpc__elasticache_subnets")), "Wrong number of elasticache_subnets returned")

	privateSubnetCidrs := outputs.StringList(t, "vpc__private_subnets_cidr_blocks")
	assert.Equal(t, []string{"10.24.8.0/23", "10.24.10.0/23", "10.24.12.0/23"}, privateSubnetCidrs, "Wrong private_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.24.0.0/24", "10.24.1.0/24", "10.24.2.0/24"}, outputs.StringList(t, "vpc__public_subnets_cidr_blocks"), "Wrong public_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.24.14.0/24", "10.24.15.0/24"}, outputs.StringList(t, "vpc__database_subnets_cidr_blocks"), "Wrong database_subnets_cidr_blocks returned")

	assert.Equal(t, privateSubnetCidrs, outputs.StringList(t, "vpc__control_subnets_cidr_blocks"), "control_subnets_cidr_blocks must equal private_subnets_cidr_blocks")
	assert.Equal(t, privateSubnetCidrs, outputs.StringList(t, "vpc__service_subnets_cidr_blocks"), "service_subnets_cidr_blocks must equal private_subnets_cidr_blocks")
	assert.Equal(t, privateSubnetCidrs, outputs.StringList(t, "vpc__compute_subnets_cidr_blocks"), "compute_subnets_cidr_blocks must equal private_subnets_cidr_blocks")
	assert.Equal(t, privateSubnets, outputs.StringList(t, "vpc__control_subnets"), "control_subnets must equal private_subnets")
	assert.Equal(t, privateSubnets, outputs.StringList(t, "vpc__service_subnets"), "service_subnets must equal private_subnets")
	assert.Equal(t, privateSubnets, outputs.StringList(t, "vpc__compute_subnets"), "compute_subnets must equal private_subnets")

	assert.Equal(t, 0, len(outputs.StringList(t, "vpc__elasticache_subnets_cidr_blocks")), "Wrong number of elasticache_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.24.4.0/23", "10.24.6.0/23"}, outputs.StringList(t, "vpc__intra_subnets_cidr_blocks"), "Wrong intra_subnets_cidr_blocks returned")

	assert.Empty(t, outputs.String(t, "vpc__vpc_ipv6_cidr_block"), "vpc_ipv6_cidr_block should be empty when ipv6 is disabled")
	assert.Equal(t, 0, len(outputs.StringList(t, "vpc__public_subnets_ipv6_cidr_blocks")), "public_subnets_ipv6_cidr_blocks should be empty when ipv6 is disabled")
	assert.Equal(t, 0, len(outputs.StringList(t, "vpc__private_subnets_ipv6_cidr_blocks")), "private_subnets_ipv6_cidr_blocks should be empty when ipv6 is disabled")

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_flow_log_id"), "vpc_flow_log_id was not returned")
	assert.Equal(t, "s3", outputs.String(t, "vpc__vpc_flow_log_destination_type"), "Flow logs must go to S3")
	flowLogBucket := outputs.String(t, "vpc__vpc_flow_log_bucket_name")
	assert.NotEmpty(t, flowLogBucket, "Flow log bucket must be created when flow_log_s3_destination_arn is not set")
	assert.Equal(t, "arn:aws:s3:::"+flowLogBucket, outputs.String(t, "vpc__vpc_flow_log_destination_arn"), "Flow logs must go to the created bucket")
	assert.Empty(t, outputs.String(t, "vpc__vpc_flow_log_cloudwatch_iam_role_arn"), "No CloudWatch IAM role should be created for S3 flow logs")
	// pri has no kms module, so the created flow log bucket uses SSE-S3
	assertFlowLogEncryption(t, e, flowLogBucket, outputs.String(t, "vpc__vpc_owner_id"), outputs.String(t, "vpc__vpc_flow_log_id"), "AES256")

	endpoints := getEndpoints(t, outputs)
	assert.ElementsMatch(t, []string{"s3"}, keys(endpoints), "Wrong vpc_endpoints returned")
	s3Policy := getEndpointPolicy(t, e, endpoints["s3"])
	assertServicePolicy(t, s3Policy, "s3", "s3:*")
	assert.NotContains(t, s3Policy, "aws:PrincipalOrgID", "s3 endpoint policy must not have an organization statement without endpoint_policy_org_id")
}

func testVpcSpoke(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	privateSubnets := outputs.StringList(t, "vpc__private_subnets")
	assert.Equal(t, 9, len(privateSubnets), "Wrong number of private_subnets returned")
	controlSubnets := outputs.StringList(t, "vpc__control_subnets")
	serviceSubnets := outputs.StringList(t, "vpc__service_subnets")
	computeSubnets := outputs.StringList(t, "vpc__compute_subnets")
	assert.Equal(t, 3, len(controlSubnets), "Wrong number of control_subnets returned")
	assert.Equal(t, 3, len(serviceSubnets), "Wrong number of service_subnets returned")
	assert.Equal(t, 3, len(computeSubnets), "Wrong number of compute_subnets returned")
	// In spoke mode the private subnets are the control, service and compute subnets in that order.
	if len(privateSubnets) == 9 && len(controlSubnets) == 3 && len(serviceSubnets) == 3 && len(computeSubnets) == 3 {
		assert.Equal(t, controlSubnets, privateSubnets[0:3], "Control and private subnet ids not matching")
		assert.Equal(t, serviceSubnets, privateSubnets[3:6], "Service and private subnet ids not matching")
		assert.Equal(t, computeSubnets, privateSubnets[6:9], "Compute and private subnet ids not matching")
	}

	assert.Equal(t, 3, len(outputs.StringList(t, "vpc__public_subnets")), "Wrong number of public_subnets returned")
	assert.Equal(t, 3, len(outputs.StringList(t, "vpc__intra_subnets")), "Wrong number of intra_subnets returned")
	databaseSubnets := outputs.StringList(t, "vpc__database_subnets")
	assert.Equal(t, 3, len(databaseSubnets), "Wrong number of database_subnets returned")
	assert.Equal(t, databaseSubnets, outputs.StringList(t, "vpc__elasticache_subnets"), "spoke elasticache_subnets must reuse database_subnets")
	assert.NotEmpty(t, outputs.String(t, "vpc__elasticache_subnet_group"), "spoke should expose an elasticache subnet group (reused database subnets)")

	assert.Equal(t, []string{"10.30.0.64/28", "10.30.0.80/28", "10.30.0.96/28", "10.30.2.0/26", "10.30.2.64/26", "10.30.2.128/26", "10.30.4.0/24", "10.30.5.0/24", "10.30.6.0/24"},
		outputs.StringList(t, "vpc__private_subnets_cidr_blocks"), "Wrong private_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.30.0.64/28", "10.30.0.80/28", "10.30.0.96/28"}, outputs.StringList(t, "vpc__control_subnets_cidr_blocks"), "Wrong control_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.30.2.0/26", "10.30.2.64/26", "10.30.2.128/26"}, outputs.StringList(t, "vpc__service_subnets_cidr_blocks"), "Wrong service_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.30.4.0/24", "10.30.5.0/24", "10.30.6.0/24"}, outputs.StringList(t, "vpc__compute_subnets_cidr_blocks"), "Wrong compute_subnets_cidr_blocks returned")
	assert.Equal(t, []string{"10.30.1.0/26", "10.30.1.64/26", "10.30.1.128/26"}, outputs.StringList(t, "vpc__public_subnets_cidr_blocks"), "Wrong public_subnets_cidr_blocks returned")
	databaseSubnetCidrs := outputs.StringList(t, "vpc__database_subnets_cidr_blocks")
	assert.Equal(t, []string{"10.30.3.0/26", "10.30.3.64/26", "10.30.3.128/26"}, databaseSubnetCidrs, "Wrong database_subnets_cidr_blocks returned")
	assert.Equal(t, databaseSubnetCidrs, outputs.StringList(t, "vpc__elasticache_subnets_cidr_blocks"), "spoke elasticache_subnets_cidr_blocks must reuse database_subnets_cidr_blocks")
	assert.Equal(t, []string{"10.30.0.0/28", "10.30.0.16/28", "10.30.0.32/28"}, outputs.StringList(t, "vpc__intra_subnets_cidr_blocks"), "Wrong intra_subnets_cidr_blocks returned")

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_ipv6_cidr_block"), "vpc_ipv6_cidr_block was not returned")
	assert.Equal(t, 3, len(outputs.StringList(t, "vpc__public_subnets_ipv6_cidr_blocks")), "Wrong number of public_subnets_ipv6_cidr_blocks returned")
	assert.Equal(t, 9, len(outputs.StringList(t, "vpc__private_subnets_ipv6_cidr_blocks")), "Wrong number of private_subnets_ipv6_cidr_blocks returned")
	assert.NotEmpty(t, outputs.StringList(t, "vpc__private_ipv6_egress_route_ids"), "private_ipv6_egress_route_ids was not returned")

	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_flow_log_id"), "vpc_flow_log_id was not returned")
	assert.Equal(t, "cloud-watch-logs", outputs.String(t, "vpc__vpc_flow_log_destination_type"), "Flow logs must go to CloudWatch by default")
	assert.NotEmpty(t, outputs.String(t, "vpc__vpc_flow_log_cloudwatch_iam_role_arn"), "CloudWatch flow log IAM role was not returned")
	assert.Empty(t, outputs.String(t, "vpc__vpc_flow_log_bucket_name"), "No flow log bucket should be created for CloudWatch flow logs")

	endpoints := getEndpoints(t, outputs)
	assert.ElementsMatch(t, []string{"sts"}, keys(endpoints), "Wrong vpc_endpoints returned, sts only must still create the endpoints module")
	assertServicePolicy(t, getEndpointPolicy(t, e, endpoints["sts"]), "sts", "sts:*")
}

// Checks the bucket's default encryption and waits for a delivered flow log file encrypted with it,
// which also proves that the bucket and key policies allow log delivery.
// With aws:kms the bucket must use the telemetry key of the kms module.
func assertFlowLogEncryption(t *testing.T, e *env.Environment, bucket string, accountId string, flowLogId string, algorithm string) {
	ctx := context.Background()
	region := e.MustRegion(t)
	client := s3.NewFromConfig(aws.Config(t, region))

	encryption, err := client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: awssdk.String(bucket)})
	require.NoError(t, err, "Failed to get the encryption of bucket %s", bucket)
	require.NotEmpty(t, encryption.ServerSideEncryptionConfiguration.Rules, "Bucket %s has no encryption rule", bucket)
	byDefault := encryption.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault
	require.Equal(t, algorithm, string(byDefault.SSEAlgorithm), "Wrong default encryption of bucket %s", bucket)
	kmsKeyArn := awssdk.ToString(byDefault.KMSMasterKeyID)
	if algorithm == "aws:kms" {
		require.NotEmpty(t, kmsKeyArn, "Bucket %s has no default KMS key", bucket)
		kmsClient := kms.NewFromConfig(aws.Config(t, region))
		aliases, err := kmsClient.ListAliases(ctx, &kms.ListAliasesInput{KeyId: awssdk.String(kmsKeyArn)})
		require.NoError(t, err, "Failed to list the aliases of KMS key %s", kmsKeyArn)
		telemetry := false
		for _, alias := range aliases.Aliases {
			t.Logf("Bucket %s uses KMS key %s with alias %s", bucket, kmsKeyArn, awssdk.ToString(alias.AliasName))
			telemetry = telemetry || strings.HasSuffix(awssdk.ToString(alias.AliasName), "/telemetry")
		}
		assert.True(t, telemetry, "Bucket %s must use the telemetry KMS key, not %s", bucket, kmsKeyArn)
		// Without this statement no flow log file is delivered, fail before waiting for one
		policy, err := kmsClient.GetKeyPolicy(ctx, &kms.GetKeyPolicyInput{KeyId: awssdk.String(kmsKeyArn), PolicyName: awssdk.String("default")})
		require.NoError(t, err, "Failed to get the policy of KMS key %s", kmsKeyArn)
		require.Contains(t, awssdk.ToString(policy.Policy), "delivery.logs.amazonaws.com",
			"KMS key %s does not allow delivery.logs.amazonaws.com, the kms module that created it is older than this vpc module", kmsKeyArn)
	} else {
		assert.Empty(t, kmsKeyArn, "Bucket %s must not have a KMS key", bucket)
	}

	// Flow logs are delivered about every 5 minutes, look at today's files only
	prefix := fmt.Sprintf("AWSLogs/%s/vpcflowlogs/%s/%s/", accountId, region, time.Now().UTC().Format("2006/01/02"))
	key, err := retry.DoWithRetryE(t, fmt.Sprintf("Wait for a flow log file in s3://%s/%s", bucket, prefix), 30, 30*time.Second, func() (string, error) {
		var newestKey string
		var newest time.Time
		pages := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{Bucket: awssdk.String(bucket), Prefix: awssdk.String(prefix)})
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				return "", err
			}
			for _, object := range page.Contents {
				if strings.HasSuffix(awssdk.ToString(object.Key), ".log.gz") && object.LastModified != nil && object.LastModified.After(newest) {
					newest = *object.LastModified
					newestKey = awssdk.ToString(object.Key)
				}
			}
		}
		if newestKey == "" {
			return "", fmt.Errorf("no flow log file in s3://%s/%s yet", bucket, prefix)
		}
		return newestKey, nil
	})
	if err != nil {
		// AWS reports why the delivery fails on the flow log itself
		flowLogs, describeErr := aws.NewEC2Client(t, region).DescribeFlowLogs(ctx, &ec2.DescribeFlowLogsInput{FlowLogIds: []string{flowLogId}})
		require.NoError(t, describeErr, "Failed to describe flow log %s", flowLogId)
		for _, flowLog := range flowLogs.FlowLogs {
			t.Logf("Flow log %s delivery status: %s %s", flowLogId, awssdk.ToString(flowLog.DeliverLogsStatus), awssdk.ToString(flowLog.DeliverLogsErrorMessage))
		}
		require.NoError(t, err, "No flow log file was delivered to bucket %s", bucket)
	}

	object, err := client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: awssdk.String(bucket), Key: awssdk.String(key)})
	require.NoError(t, err, "Failed to read flow log file s3://%s/%s", bucket, key)
	assert.Equal(t, algorithm, string(object.ServerSideEncryption), "Wrong encryption of flow log file s3://%s/%s", bucket, key)
	assert.Equal(t, kmsKeyArn, awssdk.ToString(object.SSEKMSKeyId), "Wrong KMS key of flow log file s3://%s/%s", bucket, key)
}

func getEndpoints(t *testing.T, outputs tf.Outputs) map[string]string {
	value := outputs.Map(t, "vpc__vpc_endpoints")
	result := make(map[string]string, len(value))
	for k, v := range value {
		result[k], _ = v.(string)
	}
	return result
}

func keys(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func getEndpointPolicy(t *testing.T, e *env.Environment, endpointId string) string {
	out, err := aws.NewEC2Client(t, e.MustRegion(t)).DescribeVpcEndpoints(context.Background(), &ec2.DescribeVpcEndpointsInput{VpcEndpointIds: []string{endpointId}})
	require.NoError(t, err, "Failed to describe vpc endpoint %s", endpointId)
	require.Equal(t, 1, len(out.VpcEndpoints), "vpc endpoint %s not found", endpointId)
	return awssdk.ToString(out.VpcEndpoints[0].PolicyDocument)
}

func assertServicePolicy(t *testing.T, policy string, key string, action string) {
	assert.Contains(t, policy, action, "Wrong actions in %s endpoint policy", key)
	assert.Contains(t, policy, "aws:PrincipalAccount", "%s endpoint policy must only trust the current account", key)
	assert.NotRegexp(t, `"Action"\s*:\s*(\[\s*)?"\*"`, policy, "%s endpoint policy must not allow bare * action", key)
}
