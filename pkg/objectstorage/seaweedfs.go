package objectstorage

import (
	"context"
	"encoding/json"
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
	s3SecretName       = "seaweedfs-s3-secret"
	seaweedFSNamespace = "seaweedfs-system"
	s3SecretConfigName = "seaweedfs_s3_config"
	s3AdminUser        = "anvAdmin"
)

var (
	s3Endpoint = fmt.Sprintf("http://seaweedfs-s3.%s.svc.cluster.local:8333", seaweedFSNamespace)
)

type S3Credentials struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type Credentials struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type Identity struct {
	Name        string        `json:"name"`
	Credentials []Credentials `json:"credentials"`
	Actions     []string      `json:"actions"`
}

type S3Config struct {
	Identities []Identity `json:"identities"`
}

type SeaweedFSClient struct {
	k8sClient k8sClient.Client
}

func (client *SeaweedFSClient) CreateBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error {
	s3Client, err := client.getS3Instance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get s3 client instance: %w", err)
	}

	s3BucketInput := s3.CreateBucketInput{
		Bucket: aws.String(bucket.GetName()),
	}

	_, err = s3Client.CreateBucketWithContext(ctx, &s3BucketInput)
	if err != nil {
		return fmt.Errorf("failed to create s3 bucket: %w", err)
	}
	return nil
}

func (client *SeaweedFSClient) DeleteBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error {
	return nil
}

func (client *SeaweedFSClient) UpdateBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error {
	return nil
}

func NewSeaweedFSClient(k8sClient k8sClient.Client) ObjectStorage {
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

	s3ConfigBytes, found := secret.Data[s3SecretConfigName]
	if !found {
		return nil, fmt.Errorf("secret %s does not have s3 config key %s", s3SecretName, s3SecretConfigName)
	}

	var s3Config S3Config
	err = json.Unmarshal(s3ConfigBytes, &s3Config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal s3 config entry from secret: %w", err)
	}

	var admUserIdentity *Identity
	for _, identity := range s3Config.Identities {
		if identity.Name == s3AdminUser {
			admUserIdentity = &identity
			break
		}
	}

	creds := awscredentials.NewStaticCredentials(admUserIdentity.Credentials[0].AccessKey, admUserIdentity.Credentials[0].SecretKey, "")
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
