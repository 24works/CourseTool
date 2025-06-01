package wxpush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// 微信配置變數，將從環境變數載入
var (
	appID            string
	appSecret        string
	openID           string
	courseTemplateID string
)

func init() {
	appID = os.Getenv("WXPUSH_APP_ID")
	appSecret = os.Getenv("WXPUSH_APP_SECRET")
	openID = os.Getenv("WXPUSH_OPEN_ID")
	courseTemplateID = os.Getenv("WXPUSH_COURSE_TEMPLATE_ID")

	// 檢查是否設定了所有必要的環境變數
	if appID == "" {
		log.Printf("WXPUSH 警告: 環境變數 WXPUSH_APP_ID 未設定。")
	}
	if appSecret == "" {
		log.Printf("WXPUSH 警告: 環境變數 WXPUSH_APP_SECRET 未設定。")
	}
	if openID == "" {
		log.Printf("WXPUSH 警告: 環境變數 WXPUSH_OPEN_ID 未設定。")
	}
	if courseTemplateID == "" {
		log.Printf("WXPUSH 警告: 環境變數 WXPUSH_COURSE_TEMPLATE_ID 未設定。")
	}

	if appID == "" || appSecret == "" || openID == "" || courseTemplateID == "" {
		log.Printf("WXPUSH 警告: 一個或多個 WXPUSH 環境變數未設定，相關功能可能無法正常工作。")
	}
}

// AccessTokenResponse 結構用於解析獲取 access_token 的回應
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Errcode     int    `json:"errcode"` // 微信可能在獲取 token 時也返回錯誤碼
	Errmsg      string `json:"errmsg"`  // 微信可能在獲取 token 時也返回錯誤信息
}

// SendMessageResponse 結構用於解析發送模板消息的回應
type SendMessageResponse struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	MsgID   int64  `json:"msgid"`
}

// GetAccessToken 函式用於獲取微信公眾號的 access_token
func GetAccessToken() (string, error) {
	// 在這裡再次檢查，確保在使用前變數已設定
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("獲取 access_token 失敗: WXPUSH_APP_ID 或 WXPUSH_APP_SECRET 未設定。")
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("發送請求失敗: %w", err)
	}
	defer resp.Body.Close()

	// 使用 io.ReadAll 替換 ioutil.ReadAll
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取回應失敗: %w", err)
	}

	var result AccessTokenResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("解析 JSON 回應失敗: %w", err)
	}

	// 檢查是否有錯誤碼，或者 access_token 是否為空
	if result.Errcode != 0 || result.AccessToken == "" {
		return "", fmt.Errorf("未能獲取 access_token，回應: %s", string(body))
	}

	//fmt.Printf("獲取到 access_token: %s\n", result.AccessToken)
	return result.AccessToken, nil
}

// CourseReminderData 結構用於傳遞課程提醒資訊
type CourseReminderData struct {
	CourseName     string
	TeacherName    string
	CourseLocation string
	TimeNumber     string // 例如 "第一節", "下午2點"
	NowTime        string // 當前時間，用於顯示提醒發送時間
	Note           string // 額外備註或每日一句
}

// TemplateDataValue 結構用於模板消息中的數據值
type TemplateDataValue struct {
	Value string `json:"value"`
}

// TemplateData 結構用於模板消息的數據部分
// 根據課程提醒的內容進行修改
type TemplateData struct {
	Coursename     TemplateDataValue `json:"coursename"`
	Teachername    TemplateDataValue `json:"teachername"`
	Courselocation TemplateDataValue `json:"courselocation"`
	Timenumber     TemplateDataValue `json:"timenumber"` // 例如 "第一節", "下午2點"
	Nowtime        TemplateDataValue `json:"nowtime"`    // 當前時間
	TodayNote      TemplateDataValue `json:"today_note"` // 額外備註
}

// TemplateMessage 結構用於發送模板消息的請求體
type TemplateMessage struct {
	ToUser     string       `json:"touser"`
	TemplateID string       `json:"template_id"`
	URL        string       `json:"url"`
	Data       TemplateData `json:"data"`
}

// SendCourseReminder 函式用於發送課程提醒模板消息
func SendCourseReminder(accessToken string, data CourseReminderData) error {
	// 在這裡再次檢查，確保在使用前變數已設定
	if openID == "" || courseTemplateID == "" {
		return fmt.Errorf("發送課程提醒失敗: WXPUSH_OPEN_ID 或 WXPUSH_COURSE_TEMPLATE_ID 未設定。")
	}

	// 獲取當前時間，用於 NowTime 欄位
	currentTime := time.Now().Format("2006年01月02日 15:04")

	message := TemplateMessage{
		ToUser:     openID,
		TemplateID: courseTemplateID,      // 使用新的課程提醒模板ID
		URL:        "https://www.ric.moe", // 可以替換為課程相關的連結
		Data: TemplateData{
			Coursename:     TemplateDataValue{Value: data.CourseName},
			Teachername:    TemplateDataValue{Value: data.TeacherName},
			Courselocation: TemplateDataValue{Value: data.CourseLocation},
			Timenumber:     TemplateDataValue{Value: data.TimeNumber},
			Nowtime:        TemplateDataValue{Value: currentTime},
			TodayNote:      TemplateDataValue{Value: data.Note},
		},
	}

	jsonBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化請求體失敗: %w", err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("發送請求失敗: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("讀取回應失敗: %w", err)
	}

	// 解析微信伺服器的回應
	var sendResp SendMessageResponse
	if err := json.Unmarshal(body, &sendResp); err != nil {
		// 如果解析失敗，仍然打印原始 body 並返回錯誤
		fmt.Printf("發送課程提醒回應 (解析失敗): %s\n", string(body))
		return fmt.Errorf("解析發送課程提醒回應失敗: %w, 原始回應: %s", err, string(body))
	}

	if sendResp.Errcode == 0 {
		fmt.Printf("課程提醒發送成功! msgid: %d\n", sendResp.MsgID)
	} else {
		fmt.Printf("課程提醒發送失敗! 錯誤碼: %d, 錯誤訊息: %s\n", sendResp.Errcode, sendResp.Errmsg)
		return fmt.Errorf("發送課程提醒失敗，錯誤碼: %d, 錯誤訊息: %s", sendResp.Errcode, sendResp.Errmsg)
	}

	return nil
}
