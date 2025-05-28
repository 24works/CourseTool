# --- 編譯階段 (Build Stage) ---
# 使用 golang:1.24-alpine 作為編譯環境，因為它基於 Alpine Linux，體積較小，
# 且版本滿足 go.mod 中 go >= 1.24.3 的要求。
FROM golang:1.24-alpine AS builder

# 設定工作目錄為 /app
WORKDIR /app

# 複製整個 src 目錄到容器中
# 根據您的說明，go.mod 和 go.sum 檔案現在位於 src 目錄內。
COPY src ./src

# 將工作目錄切換到 Go 模組的根目錄，即 /app/src
# 這是為了讓 go mod download 和 go build 能夠在正確的 Go 模組上下文中執行。
WORKDIR /app/src

# 偵錯步驟：列出當前目錄（/app/src）的內容，以確認檔案結構
RUN ls -R .

# 下載 Go 模組
# 這裡使用 --mount=type=cache 來快取 Go 模組，進一步提高構建速度
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download

# 編譯 Go 程式
# -o CourseTool 指定輸出二進位檔案的名稱為 CourseTool
# . 指向當前工作目錄 (即 /app/src)，這是 Go 模組的根目錄
# CGO_ENABLED=0 禁用 CGO，這使得編譯出的二進位檔案完全靜態連結，不依賴於系統函式庫，
# 適合在極小的基礎映像（如 scratch 或 alpine）中運行。
# -ldflags="-s -w" 用於移除符號表和調試資訊，進一步縮小二進位檔案體積。
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o CourseTool .

# --- 運行階段 (Run Stage) ---
# 使用 alpine:latest 作為最終運行映像，這是非常小的 Linux 發行版。
FROM alpine:latest

# 安裝 ca-certificates，用於處理 HTTPS 連線
RUN apk --no-cache add ca-certificates

# 安裝 tzdata 並設定時區
# 您可以將 "Asia/Taipei" 替換為您需要的時區，例如 "Asia/Shanghai", "America/New_York", "Europe/London" 等。
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && \
    echo "Asia/Taipei" > /etc/timezone

# 也可以直接設定 ENV TZ 環境變數，Go 程式通常會優先讀取此變數
ENV TZ="Asia/Taipei"

# 設定工作目錄
WORKDIR /root/

# 從編譯階段複製編譯好的 CourseTool 二進位檔案
# 現在 CourseTool 在編譯階段的 /app/src/CourseTool
COPY --from=builder /app/src/CourseTool .

# 設定容器啟動時執行的命令
# 這裡直接執行 CourseTool 程式
ENTRYPOINT ["./CourseTool"]

# 可以選擇性地暴露程式監聽的埠，如果您的程式是一個網路服務
# EXPOSE 8080
