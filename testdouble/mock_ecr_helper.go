package testdouble

import (
	"context"
	"errors"

	// "fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func GenerateMockECRAPI(mockParams MockECRParams) MockECRAPI {
    return MockECRAPI{
        ListImagesAPI:      GenerateMockECRListImagesAPI(mockParams),
        DescribeImagesAPI:  GenerateMockECRDescribeImagesAPI(mockParams),
    }
}

func GenerateMockECRListImagesAPI(mockParams MockECRParams) MockECRListImagesAPI {
    return MockECRListImagesAPI(func(ctx context.Context, params *ecr.ListImagesInput, optFns ...func(*ecr.Options)) (*ecr.ListImagesOutput, error) {
        // fmt.Printf("MockECRListImagesAPI(Expect) : %d / %s / %s\n", mockParams.ECRParams.MaxResults, mockParams.ECRParams.RegistryId, mockParams.ECRParams.RepositoryName)
        // fmt.Printf("MockECRListImagesAPI(Real) :   %d / %s / %s\n", aws.ToInt32(params.MaxResults), aws.ToString(params.RegistryId), aws.ToString(params.RepositoryName))

        if params.MaxResults == nil || aws.ToInt32(params.MaxResults) != mockParams.ECRParams.MaxResults {
            return nil, errors.New("ListImagesを呼び出すときのMaxResultsの指定が間違っています")
        }
        if params.RegistryId == nil || aws.ToString(params.RegistryId) != mockParams.ECRParams.RegistryId {
            return nil, errors.New("ListImagesを呼び出すときのRegistryIdの指定が間違っています")
        }
        if params.RepositoryName == nil || aws.ToString(params.RepositoryName) != mockParams.ECRParams.RepositoryName {
            return nil, errors.New("ListImagesを呼び出すときのRepositoryNameの指定が間違っています")
        }
        imagesOutput := &ecr.ListImagesOutput {
            ImageIds:   mockParams.ECRParams.ImageIds,
        }
        return imagesOutput, nil
    })
}

func GenerateMockECRDescribeImagesAPI(mockParams MockECRParams) MockECRDescribeImagesAPI {
    return MockECRDescribeImagesAPI(func(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
        // fmt.Printf("MockECRDescribeImagesAPI(Expect) : %d / %s / %s\n", mockParams.ECRParams.MaxResults, mockParams.ECRParams.RegistryId, mockParams.ECRParams.RepositoryName)
        // fmt.Printf("MockECRDescribeImagesAPI(Real) :   %d / %s / %s\n", aws.ToInt32(params.MaxResults), aws.ToString(params.RegistryId), aws.ToString(params.RepositoryName))

        if params.MaxResults == nil || aws.ToInt32(params.MaxResults) != mockParams.ECRParams.MaxResults {
            return nil, errors.New("DescribeImagesを呼び出すときのMaxResultsの指定が間違っています")
        }
        if params.RegistryId == nil || aws.ToString(params.RegistryId) != mockParams.ECRParams.RegistryId {
            return nil, errors.New("DescribeImagesを呼び出すときのRegistryIdの指定が間違っています")
        }
        if params.RepositoryName == nil || aws.ToString(params.RepositoryName) != mockParams.ECRParams.RepositoryName {
            return nil, errors.New("DescribeImagesを呼び出すときのRepositoryNameの指定が間違っています")
        }
        detailOutput := &ecr.DescribeImagesOutput {
            ImageDetails:   mockParams.ECRParams.ImageDetails,
        }
        return detailOutput, nil
    })
}