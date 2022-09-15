package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// 対象のイメージタグを検索
func ImageTag(ids []types.ImageIdentifier, tags []string, digest string) string {
    for _, id := range ids {
        for _, tag := range tags {
            // nil pointer 回避
            if id.ImageTag != nil {
                // 値の比較を可能にするために変数に代入
                vsTag := *id.ImageTag
                vsDigest := *id.ImageDigest
                if vsTag == tag && vsDigest == digest {
                    return tag
                }
            }
        }
    }
    return ""
}

// ECR クライアント生成
func EcrClient(region string) (*ecr.Client, error) {
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("AWS（API）の認証に失敗しました : %s", err)
    }
    return ecr.NewFromConfig(cfg), nil
}

// ECR ListImages
type EcrListImagesAPI interface {
    ListImages(ctx context.Context, params *ecr.ListImagesInput, optFns ...func(*ecr.Options)) (*ecr.ListImagesOutput, error)
}

func EcrListImages(ctx context.Context, api EcrListImagesAPI, repositoryName string, registryId string) ([]types.ImageIdentifier, error) {
    // ページネーションさせないために最大件数を 1,000 に（実際には数十個程度の想定）
    maxResults := int32(1000)

    ecrImageIds, err := api.ListImages(ctx, &ecr.ListImagesInput{
        RepositoryName: &repositoryName,
        RegistryId:     &registryId,
        MaxResults:     &maxResults,
    })
    if err != nil {
        return nil, fmt.Errorf("リポジトリ（%s）のイメージ一覧の取得に失敗しました : %s", repositoryName, err)
    }
    return ecrImageIds.ImageIds, nil
}

// ECR DescribeImages
type EcrDescribeImagesAPI interface {
    DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error)
}

func EcrDescribeImages(ctx context.Context, api EcrDescribeImagesAPI, repositoryName string, registryId string) ([]types.ImageDetail, error) {
    // ページネーションさせないために最大件数を 1,000 に（実際には数十個程度の想定）
    maxResults := int32(1000)

    ecrImages, err := api.DescribeImages(ctx, &ecr.DescribeImagesInput{
        RepositoryName: &repositoryName,
        RegistryId:     &registryId,
        MaxResults:     &maxResults,
    })
    if err != nil {
        return nil, fmt.Errorf("リポジトリ（%s）のイメージ詳細一覧の取得に失敗しました : %s", repositoryName, err)
    }
    return ecrImages.ImageDetails, nil
}

// ECR リポジトリ内イメージ一覧取得
func ImageList(repositoryName string, registryId string, repositoryUri string) ([]Image, error) {
    var err error
    region := strings.Split(repositoryUri, ".")[3]
    ecrClient, err := EcrClient(region)
    if (err != nil) {
        return nil, err
    }

    imageIds, err := EcrListImages(context.TODO(), ecrClient, repositoryName, registryId)
    if (err != nil) {
        return nil, err
    }
    imageDetails, err := EcrDescribeImages(context.TODO(), ecrClient, repositoryName, registryId)
    if (err != nil) {
        return nil, err
    }

    var imageList []Image
    for _, v := range imageDetails {
        digest := v.ImageDigest
        pushedAt := v.ImagePushedAt
        size := v.ImageSizeInBytes
        tags := v.ImageTags
        // URI に使われているタグを検索
        tag := ImageTag(imageIds, tags, *digest)
        var uri string
        if tag == "" {
            // タグがない場合はダイジェスト
            uri = fmt.Sprintf("%s@%s", repositoryUri, *digest)
        } else {
            uri = fmt.Sprintf("%s:%s", repositoryUri, tag)
        }
        image := Image{
            Digest:         *digest,
            PushedAt:       *pushedAt,
            RepositoryName: repositoryName,
            Size:           float32(*size),
            Tags:           tags,
            Uri:            uri,
        }
        imageList = append(imageList, image)
    }
    return imageList, nil
}
