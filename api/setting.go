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
    var err error
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
    oldSettingItems, err := ReadSettingFromFile(oldSettingFile)
    if err == nil {
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
func UpdateSetting(settingPath string, cronPath string, cronCmd string, cronLog string, serviceName string, setting *Setting) error {
    var err error
    settingFile := fmt.Sprintf("%s/%s-release-setting", settingPath, serviceName)
    f, err := os.Create(settingFile)
    if err != nil {
        return err
    }
    defer f.Close()

    imageUri := *setting.ImageUri
    tmpReleaseAt := *setting.ReleaseAt
    tmpNow := time.Now().In(time.Local).Add(1 * time.Minute)
    now := time.Date(tmpNow.Year(), tmpNow.Month(), tmpNow.Day(), tmpNow.Hour(), tmpNow.Minute(), 0, 0, time.Local)
    if tmpReleaseAt.Before(now) {
        tmpReleaseAt = now
        *setting.ReleaseAt = now
    }
    releaseAt := tmpReleaseAt.Format("2006-01-02T15:04:05Z07:00")
    _, err = f.WriteString(fmt.Sprintf("%s\n%s", imageUri, releaseAt))
    if err != nil || cronCmd == "" {
        return err
    }
    // cron.d にリリーススクリプト起動用の設定を保存
    cronFile := fmt.Sprintf("%s%s-release", cronPath, serviceName)
    fmt.Printf("保存先ファイル名 : %s\n", cronFile)
    fc, err := os.Create(cronFile)
    if err != nil {
        return err
    }
    defer fc.Close()

    month := int(tmpReleaseAt.Local().Month())
    day := tmpReleaseAt.Local().Day()
    hour := tmpReleaseAt.Local().Hour()
    minute := tmpReleaseAt.Local().Minute()
    cronTime := fmt.Sprintf("%d %d %d %d * ", minute, hour, day, month)
    cronMain := fmt.Sprintf("root flock %s/%s-release-processing %s %s %s %s && ", settingPath, serviceName, cronCmd, imageUri, serviceName, cronLog)
    cronAfter1 := fmt.Sprintf("mv -f %s/%s-release-setting %s/%s-released && rm -f %s%s-release && ", settingPath, serviceName, settingPath, serviceName, cronPath, serviceName)
    cronAfter2 := fmt.Sprintf("rm -f %s/%s-release-processing\n", settingPath, serviceName)
    _, err = fc.WriteString(cronTime + cronMain + cronAfter1 + cronAfter2)

    return err
}
