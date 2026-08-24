:: get folder name without path
for /f "delims=" %%A in ('cd') do (
    set foldername=%%~nxA
)

echo. Current Folder Name: %foldername%
rm %foldername%

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -ldflags "-s -w" -o %foldername%
:: comment this if you don't have upx or don't want to use upx
:: upx for compress the binary size
D:\upx-4.2.2-win64\upx.exe -3 -q %foldername%