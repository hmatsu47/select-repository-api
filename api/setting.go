package api

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type SettingItems struct {
    ImageUri  string
    ReleaseAt time.Time
}

// 指定ファイルから設定を取得
func ReadSettingFromFile(settingFile string) (*SettingItems, error) {
    f, err := os.Open(settingFile)
    if err != nil {
        return nil, fmt.Errorf("ファイルがありません : %s", settingFile)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    scanner.Scan()
    imageUri := scanner.Text()
    scanner.Scan()
    tmpReleaseAt := scanner.Text()
    releaseAt, err := time.Parse("2006-01-02T15:04:05Z07:00", tmpReleaseAt)
    if err != nil {
        return nil, fmt.Errorf("リリース日時の形式が誤っています : %s", tmpReleaseAt)
    }
    return &SettingItems{
        ImageUri:  imageUri,
        ReleaseAt: releaseAt,
    }, nil
}

// 設定読み込み
func ReadSetting(settingPath string, serviceName string) Setting {
    // リリース設定ファイルがあればその情報を返す
    settingFile := fmt.Sprintf("%s/%s-release-setting", settingPath, serviceName)
    settingItems, err := ReadSettingFromFile(settingFile)
    if err == nil {
        return Setting{
            ImageUri:   &settingItems.ImageUri,
            IsReleased: false,
            ReleaseAt:  &settingItems.ReleaseAt,
        }
    }

    // リリース済みの設定ファイルがあればその情報を返す
    oldSettingFile := fmt.Sprintf("%s/%s-released", settingPath, serviceName)
    oldSettingItems, oerr := ReadSettingFromFile(oldSettingFile)
    if oerr == nil {
        return Setting{
            ImageUri:   &oldSettingItems.ImageUri,
            IsReleased: true,
            ReleaseAt:  &oldSettingItems.ReleaseAt,
        }
    }

    // どちらも存在しなければ「IsReleased: false」のみ返す
    return Setting{
        IsReleased: false,
    }
}

// リリース処理中かどうか確認
func CheckNowReleaseProcessing(settingPath string, serviceName string) bool {
    processingFile := fmt.Sprintf("%s/%s-release-processing", settingPath, serviceName)
    _, err := os.Stat(processingFile)
    return err == nil
}

// 設定書き込み（上書き）
func UpdateSetting(settingPath string, serviceName string, setting *Setting) error {
    settingFile := fmt.Sprintf("%s/%s-release-setting", settingPath, serviceName)
    f, err := os.Create(settingFile)
    if err != nil {
        return err
    }
    defer f.Close()

    imageUri := *setting.ImageUri
    tmpReleaseAt := *setting.ReleaseAt
    releaseAt := tmpReleaseAt.Format("2006-01-02T15:04:05Z07:00")
    _, werr := f.WriteString(fmt.Sprintf("%s\n%s", imageUri, releaseAt))

    return werr
}
