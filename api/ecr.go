package api

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// ECR API interface
type ECRAPI interface {
	EcrListImagesAPI
	EcrDescribeImagesAPI
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
		RepositoryName: aws.String(repositoryName),
		RegistryId:     aws.String(registryId),
		MaxResults:     aws.Int32(maxResults),
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
		RepositoryName: aws.String(repositoryName),
		RegistryId:     aws.String(registryId),
		MaxResults:     aws.Int32(maxResults),
	})
	if err != nil {
		return nil, fmt.Errorf("リポジトリ（%s）のイメージ詳細一覧の取得に失敗しました : %s", repositoryName, err)
	}
	return ecrImages.ImageDetails, nil
}

// ImageList を取得
func GetImageList(imageIds []types.ImageIdentifier, imageDetails []types.ImageDetail, repositoryName string, repositoryUri string) []Image {
	var imageList []Image
	for _, v := range imageDetails {
		digest := v.ImageDigest
		pushedAt := v.ImagePushedAt
		size := v.ImageSizeInBytes
		tags := v.ImageTags
		// URI に使われているタグを検索
		tag := ImageTag(imageIds, tags, aws.ToString(digest))
		var uri string
		if tag == "" {
			// タグがない場合はダイジェスト
			uri = fmt.Sprintf("%s@%s", repositoryUri, aws.ToString(digest))
		} else {
			uri = fmt.Sprintf("%s:%s", repositoryUri, tag)
		}
		image := Image{
			Digest:         aws.ToString(digest),
			PushedAt:       aws.ToTime(pushedAt),
			RepositoryName: repositoryName,
			Size:           float32(aws.ToInt64(size)),
			Tags:           tags,
			Uri:            uri,
		}
		imageList = append(imageList, image)
	}
	// 結果をプッシュ時間の降順でソート
	sort.Slice(imageList, func(i, j int) bool {
		return imageList[i].PushedAt.After(imageList[j].PushedAt)
	})
	return imageList
}

// ECR リポジトリ内イメージ一覧取得
func ImageList(ctx context.Context, api ECRAPI, repositoryName string, registryId string, repositoryUri string) ([]Image, error) {
	imageIds, err := EcrListImages(context.TODO(), api, repositoryName, registryId)
	if err != nil {
		return nil, err
	}
	imageDetails, err := EcrDescribeImages(context.TODO(), api, repositoryName, registryId)
	if err != nil {
		return nil, err
	}

	imageList := GetImageList(imageIds, imageDetails, repositoryName, repositoryUri)
	return imageList, nil
}
