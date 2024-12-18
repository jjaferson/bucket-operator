package objectstorage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awscredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	objectstoragev1alpha1 "mystorage.sh/bucket/api/v1alpha1"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	s3SecretName           = "seaweedfs-s3-secret"
	s3AdminAccessKeyId     = "admin_access_key_id"
	s3AdminSecretAccessKey = "admin_secret_access_key"
	seaweedFSNamespace     = "seaweedfs-system"
	s3AdminUser            = "anvAdmin"
)

var (
	s3Endpoint = fmt.Sprintf("http://seaweedfs-s3.%s.svc.cluster.local:8333", seaweedFSNamespace)
)

type SeaweedFSClient struct {
	k8sClient k8sClient.Client
}

func (client *SeaweedFSClient) CreateBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error {
	s3Client, err := client.getS3Instance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get s3 client instance: %w", err)
	}

	exists, err := checksBucketExists(ctx, s3Client, bucket.GetName())
	if err != nil {
		return fmt.Errorf("failed to create s3 bucket: %w", err)
	}

	if !exists {
		_, err = s3Client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucket.GetName()),
		})
		if err != nil {
			return fmt.Errorf("failed to create s3 bucket: %w", err)
		}
	}

	return nil
}

func (client *SeaweedFSClient) DeleteBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error {
	s3Client, err := client.getS3Instance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get s3 client instance: %w", err)
	}

	exists, err := checksBucketExists(ctx, s3Client, bucket.GetName())
	if err != nil {
		return fmt.Errorf("failed to delete s3 bucket: %w", err)
	}

	if exists {
		_, err = s3Client.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucket.GetName()),
		})
		if err != nil {
			return fmt.Errorf("failed to delete s3 bucket: %w", err)
		}
	}

	return nil
}

func checksBucketExists(ctx context.Context, s3Client *s3.S3, bucketName string) (bool, error) {

	bucketList, err := s3Client.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return false, fmt.Errorf("failed to list s3 bucket: %w", err)
	}

	for i := range bucketList.Buckets {
		bucket := bucketList.Buckets[i]
		if *bucket.Name == bucketName {
			return true, nil
		}
	}

	return false, nil
}

func NewSeaweedFSClient(k8sClient k8sClient.Client) *SeaweedFSClient {
	return &SeaweedFSClient{
		k8sClient: k8sClient,
	}
}

func (client *SeaweedFSClient) getS3Instance(ctx context.Context) (*s3.S3, error) {
	secret := v1.Secret{}
	err := client.k8sClient.Get(ctx, types.NamespacedName{Namespace: seaweedFSNamespace, Name: s3SecretName}, &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret %s with s3 credential: %w", s3SecretName, err)
	}

	//TODO: get aws cred from volume mount
	awsId, found := secret.Data[s3AdminAccessKeyId]
	if !found {
		return nil, fmt.Errorf("secret %s does not have aws key %s", s3SecretName, s3AdminAccessKeyId)
	}
	awsSecret, found := secret.Data[s3AdminSecretAccessKey]
	if !found {
		return nil, fmt.Errorf("secret %s does not have aws secret key %s", s3SecretName, s3AdminSecretAccessKey)
	}

	creds := awscredentials.NewStaticCredentials(string(awsId), string(awsSecret), "")
	awsSess, err := awssession.NewSession(&aws.Config{
		Credentials:      creds,
		Endpoint:         aws.String(s3Endpoint),
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return s3.New(awsSess), nil
}
