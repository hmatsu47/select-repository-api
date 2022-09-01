package api

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// ECR リポジトリ内イメージ一覧取得
func ImageList(repositoryName string, registryId string, repositoryUri string) ([]Image, error) {
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
    if err != nil {
        return nil, fmt.Errorf("AWS（API）の認証に失敗しました : %s", err)
    }

    ecrImagesClient := ecr.NewFromConfig(cfg)
    ecrImages, eerr := ecrImagesClient.DescribeImages(context.TODO(), &ecr.DescribeImagesInput{
        RepositoryName: &repositoryName,
        RegistryId: &registryId,
    })
    if eerr != nil {
        return nil, fmt.Errorf("リポジトリ（%s）のイメージ一覧の取得に失敗しました : %s", repositoryName, eerr)
    }

    imageDetails := ecrImages.ImageDetails

    var imageList []Image
    for _, v := range imageDetails {
        digest := v.ImageDigest
        pushedAt := v.ImagePushedAt
        size := v.ImageSizeInBytes
        tags := v.ImageTags
        uri := fmt.Sprintf("%s:%s", repositoryUri, v.ImageTags[0])
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
