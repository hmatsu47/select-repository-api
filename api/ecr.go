package api

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// 対象のイメージタグを検索
func ImageTag(ids []types.ImageIdentifier, tags []string, digest string) (*string, error) {
    for _, id := range ids {
        for _, tag := range tags {
            // 値の比較を可能にするために変数に代入
            vsTag := *id.ImageTag
            vsDigest := *id.ImageDigest
            if vsTag == tag && vsDigest == digest {
                return &tag, nil
            }
        }
    }
    return nil, fmt.Errorf("対象のイメージタグが見つかりません")
}

// ECR リポジトリ内イメージ一覧取得
func ImageList(repositoryName string, registryId string, repositoryUri string) ([]Image, error) {
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
    if err != nil {
        return nil, fmt.Errorf("AWS（API）の認証に失敗しました : %s", err)
    }

    // ページネーションさせないために最大件数を 1,000 に（実際には数十個程度の想定）
    maxResults := int32(1000)
    ecrImagesClient := ecr.NewFromConfig(cfg)

    // 一旦イメージ一覧を取得しておく（URI の一部としてどのタグを使っているのかを後で検索する）
    ecrImageIds, ierr := ecrImagesClient.ListImages(context.TODO(), &ecr.ListImagesInput{
        RepositoryName: &repositoryName,
        RegistryId: &registryId,
        MaxResults: &maxResults,
    })
    if ierr != nil {
        return nil, fmt.Errorf("リポジトリ（%s）のイメージ一覧の取得に失敗しました : %s", repositoryName, ierr)
    }
    imageIds := ecrImageIds.ImageIds

    // イメージ詳細一覧を取得
    ecrImages, eerr := ecrImagesClient.DescribeImages(context.TODO(), &ecr.DescribeImagesInput{
        RepositoryName: &repositoryName,
        RegistryId: &registryId,
        MaxResults: &maxResults,
    })
    if eerr != nil {
        return nil, fmt.Errorf("リポジトリ（%s）のイメージ詳細一覧の取得に失敗しました : %s", repositoryName, eerr)
    }

    imageDetails := ecrImages.ImageDetails

    var imageList []Image
    for _, v := range imageDetails {
        digest := v.ImageDigest
        pushedAt := v.ImagePushedAt
        size := v.ImageSizeInBytes
        tags := v.ImageTags
        // URI に使われているタグを検索
        tag, terr := ImageTag(imageIds, tags, *digest)
        if terr != nil {
            return nil, fmt.Errorf("リポジトリ（%s）のイメージURIの取得に失敗しました : %s", repositoryName, terr)
        }
        uri := fmt.Sprintf("%s:%s", repositoryUri, *tag)
        image := Image{
            Digest: *digest,
            PushedAt: *pushedAt,
            RepositoryName: repositoryName,
            Size: float32(*size),
            Tags: tags,
            Uri: uri,
        }
        imageList = append(imageList, image)
    }
    return imageList, nil
}
