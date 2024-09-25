package objectstorage

import (
	"context"

	objectstoragev1alpha1 "mystorage.sh/bucket/api/v1alpha1"
)

type ObjectStorage interface {
	CreateBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error
	DeleteBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error
	UpdateBucket(ctx context.Context, bucket *objectstoragev1alpha1.Bucket) error
}
