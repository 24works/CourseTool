package update // 將 package main 改為 package update

import (
	ASNIColor "CourseTool/asnicolor" // 新增：引入 ASNIColor 包
	"fmt"
	"io" // For io.Copy and io.ReadAll
	"log"
	"net/http"
	"os"            // For os.Create, os.Executable, os.Remove, os.Rename, os.Stat, os.ReadFile
	"path/filepath" // For getting executable path
	"runtime"       // For OS detection
	"strconv"
	"strings"
	"sync" // 新增：用於 ProgressBarWriter 的互斥鎖
)

// CurrentAppVersion 定義當前應用程式的版本
// 這個版本號應該與您 main.go 中橫幅顯示的版本一致
const CurrentAppVersion = "1.0.0"

// ProgressBarWriter 是一個 io.Writer，用於顯示下載進度條
type ProgressBarWriter struct {
	writer       io.Writer  // 底層的文件寫入器
	total        int64      // 總字節數
	downloaded   int64      // 已下載的字節數
	lastProgress int        // 上次打印的百分比，用於控制打印頻率
	mu           sync.Mutex // 保護 downloaded 和 lastProgress
}

// NewProgressBarWriter 創建一個新的 ProgressBarWriter 實例
func NewProgressBarWriter(w io.Writer, total int64) *ProgressBarWriter {
	return &ProgressBarWriter{
		writer: w,
		total:  total,
	}
}

// Write 實現 io.Writer 接口。每次寫入數據時，它會更新已下載的字節數，並根據進度打印進度條。
func (pbw *ProgressBarWriter) Write(p []byte) (n int, err error) {
	// 將數據寫入到底層的文件寫入器
	n, err = pbw.writer.Write(p)
	if err != nil {
		return
	}

	pbw.mu.Lock() // 鎖定以保護共享狀態
	pbw.downloaded += int64(n)

	if pbw.total > 0 {
		// 計算當前進度百分比
		currentProgress := int((float64(pbw.downloaded) / float64(pbw.total)) * 100)

		// 每隔 5% 打印一次進度，或者在 0% 和 100% 時打印
		if currentProgress > pbw.lastProgress || currentProgress == 0 || currentProgress == 100 {
			pbw.printProgress(currentProgress)
			pbw.lastProgress = currentProgress
		}
	} else {
		// 如果無法獲取總大小，則只顯示已下載的字節數
		fmt.Printf("\rDownloading: %s downloaded...", ASNIColor.BrightCyan+byteCountToHuman(pbw.downloaded)+ASNIColor.Reset)
	}
	pbw.mu.Unlock() // 解鎖

	return
}

// printProgress 打印進度條到控制台
func (pbw *ProgressBarWriter) printProgress(progress int) {
	const barLength = 20 // 進度條的長度
	filled := int(float64(progress) / 100 * barLength)
	empty := barLength - filled

	// 構建進度條的填充部分和空白部分
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", empty)

	// 使用 \r 實現光標回車，以便在同一行更新進度條
	fmt.Printf("\rDownloading: %s%3d%%%s [%s%s%s]",
		ASNIColor.BrightCyan, progress, ASNIColor.Reset, // 百分比
		ASNIColor.BrightGreen, bar, ASNIColor.Reset) // 進度條視覺部分

	// 如果下載完成，換行以避免後續輸出覆蓋進度條
	if progress == 100 {
		fmt.Println()
	}
}

// byteCountToHuman 將字節數轉換為人類可讀的格式 (例如 1.2 MB)
func byteCountToHuman(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// getRemoteVersion 從指定的 URL 獲取遠端版本號
func getRemoteVersion(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("獲取遠端版本失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("獲取遠端版本時收到非 200 狀態碼: %d %s", resp.StatusCode, resp.Status)
	}

	// 將 ioutil.ReadAll 替換為 io.ReadAll
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取遠端版本響應失敗: %v", err)
	}

	// 移除可能存在的換行符或空白字元
	remoteVersion := strings.TrimSpace(string(body))
	return remoteVersion, nil
}

// parseVersion 將版本字串解析為整數切片 (例如 "1.2.3" -> [1, 2, 3])
func parseVersion(version string) ([]int, error) {
	parts := strings.Split(version, ".")
	var intParts []int
	for _, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("解析版本號部分 '%s' 失敗: %v", part, err)
		}
		intParts = append(intParts, val)
	}
	return intParts, nil
}

// compareVersions 比較兩個版本字串
// 如果 versionA > versionB 返回 1
// 如果 versionA < versionB 返回 -1
// 如果 versionA == versionB 返回 0
func compareVersions(versionA, versionB string) (int, error) {
	vA, err := parseVersion(versionA)
	if err != nil {
		return 0, fmt.Errorf("解析版本 A 失敗: %v", err)
	}
	vB, err := parseVersion(versionB)
	if err != nil {
		return 0, fmt.Errorf("解析版本 B 失敗: %v", err)
	}

	maxLength := len(vA)
	if len(vB) > maxLength {
		maxLength = len(vB)
	}

	for i := 0; i < maxLength; i++ {
		valA := 0
		if i < len(vA) {
			valA = vA[i]
		}
		valB := 0
		if i < len(vB) {
			valB = vB[i]
		}

		if valA > valB {
			return 1, nil
		}
		if valA < valB {
			return -1, nil
		}
	}
	return 0, nil // 版本相等
}

// downloadFile 下載文件到指定路徑
func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("創建文件失敗: %v", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下載文件失敗: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下載文件時收到非 200 狀態碼: %d %s", resp.StatusCode, resp.Status)
	}

	var writer io.Writer = out
	totalSize := resp.ContentLength // 從響應頭獲取文件總大小

	if totalSize > 0 {
		// 如果能獲取到總大小，則使用 ProgressBarWriter
		pbw := NewProgressBarWriter(out, totalSize)
		writer = pbw
		fmt.Println("Starting download...") // 初始提示
	} else {
		// 如果無法獲取總大小，則只顯示已下載的字節數
		fmt.Println("Starting download (total size unknown)...")
	}

	// 將響應體複製到寫入器，這會觸發 ProgressBarWriter 的 Write 方法
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("寫入文件失敗: %v", err)
	}
	return nil
}

// CheckForUpdates 檢查是否有新的應用程式版本可用並在 Windows 上執行更新
func CheckForUpdates() {
	remoteVersionURL := "https://coursetool.ric.moe/CTversion"          // 遠端版本資訊的 URL
	downloadURL := "https://software.ric.moe/CourseTool/CourseTool.exe" // Windows 更新下載 URL

	log.Printf("正在檢查更新... 當前版本: %s\n", CurrentAppVersion)

	remoteVersion, err := getRemoteVersion(remoteVersionURL)
	if err != nil {
		log.Printf("檢查更新失敗: %v\n", err)
		return
	}

	log.Printf("遠端最新版本: %s\n", remoteVersion)

	comparison, err := compareVersions(CurrentAppVersion, remoteVersion)
	if err != nil {
		log.Printf("比較版本失敗: %v\n", err)
		return
	}

	if comparison == -1 {
		// 使用 ASNIColor 進行輸出
		fmt.Printf("***** %s有新版本可用！*****\n", ASNIColor.BrightGreen+"更新提醒"+ASNIColor.Reset)

		if runtime.GOOS == "windows" {
			fmt.Printf("檢測到 Windows 系統，正在嘗試自動更新...\n")

			// 獲取當前執行檔的路徑
			exePath, err := os.Executable() // os.Executable 已經是推薦的替代方案
			if err != nil {
				log.Printf("獲取當前執行檔路徑失敗: %v\n", err)
				return
			}
			exeDir := filepath.Dir(exePath)
			exeName := filepath.Base(exePath)

			tempFileName := filepath.Join(exeDir, exeName+".new")
			oldFileName := filepath.Join(exeDir, exeName+".old")

			fmt.Printf("正在下載新版本到: %s\n", tempFileName)
			err = downloadFile(tempFileName, downloadURL)
			if err != nil {
				log.Printf("下載新版本失敗: %v\n", err)
				return
			}
			fmt.Printf("新版本下載完成。\n")

			// 嘗試將舊的執行檔重命名
			// 如果 oldFileName 已經存在，先刪除它（可能是上次更新失敗留下的）
			if _, err := os.Stat(oldFileName); err == nil { // os.Stat 已經是推薦的替代方案
				if err := os.Remove(oldFileName); err != nil { // os.Remove 已經是推薦的替代方案
					log.Printf("刪除舊的備份執行檔失敗: %v\n", err)
					// 不返回，嘗試繼續
				}
			}

			fmt.Printf("正在備份舊版本 (%s) 到 (%s)...\n", exePath, oldFileName)
			err = os.Rename(exePath, oldFileName) // os.Rename 已經是推薦的替代方案
			if err != nil {
				log.Printf("備份舊版本失敗: %v\n", err)
				// 如果備份失敗，嘗試清理下載的新文件
				os.Remove(tempFileName)
				return
			}
			fmt.Printf("舊版本備份完成。\n")

			// 將新下載的文件重命名為當前執行檔名稱
			fmt.Printf("正在用新版本覆蓋舊版本...\n")
			err = os.Rename(tempFileName, exePath) // os.Rename 已經是推薦的替代方案
			if err != nil {
				log.Printf("覆蓋舊版本失敗: %v\n", err)
				// 如果覆蓋失敗，嘗試恢復舊版本
				if err := os.Rename(oldFileName, exePath); err != nil {
					log.Printf("恢復舊版本失敗: %v\n", err)
				}
				return
			}
			fmt.Printf("%s更新成功！請重新啟動程式以應用新版本。\n", ASNIColor.BrightGreen+"自動更新"+ASNIColor.Reset)

			// 嘗試刪除舊的備份文件
			if err := os.Remove(oldFileName); err != nil { // os.Remove 已經是推薦的替代方案
				log.Printf("刪除舊版本備份文件失敗: %v\n", err)
			}

		} else {
			fmt.Printf("請訪問 %s 下載最新版本並手動更新。\n", "https://github.com/RichardMiku/CourseTool")
		}
	} else if comparison == 1 {
		fmt.Printf("%s 您當前版本 (%s) 比遠端版本 (%s) 更新。這可能是開發版本或錯誤。\n", ASNIColor.Yellow+"警告:"+ASNIColor.Reset, CurrentAppVersion, remoteVersion)
	} else {
		fmt.Printf("%s 您當前版本是最新的 (%s)。\n", ASNIColor.BrightGreen+"已是最新版本"+ASNIColor.Reset, CurrentAppVersion)
	}
}
