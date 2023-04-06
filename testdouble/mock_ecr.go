package testdouble

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// モックパラメーター
type ECRParams struct {
    RepositoryName  string
    RegistryId      string
    ImageIds        []types.ImageIdentifier
    ImageDetails    []types.ImageDetail
    MaxResults      int32
}

// モック生成用
type MockECRParams struct {
    ECRParams       ECRParams
}

// モック化
type MockECRAPI struct {
    ListImagesAPI       MockECRListImagesAPI
    DescribeImagesAPI   MockECRDescribeImagesAPI
}

type MockECRListImagesAPI       func(ctx context.Context, params *ecr.ListImagesInput, optFns ...func(*ecr.Options)) (*ecr.ListImagesOutput, error)
type MockECRDescribeImagesAPI   func(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error)

func (m MockECRAPI) ListImages(ctx context.Context, params *ecr.ListImagesInput, optFns ...func(*ecr.Options)) (*ecr.ListImagesOutput, error) {
    return m.ListImagesAPI(ctx, params, optFns...)
}

func (m MockECRAPI) DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
    return m.DescribeImagesAPI(ctx, params, optFns...)
}