
:: get folder name without path
for /f "delims=" %%A in ('cd') do (
    set foldername=%%~nxA
)

echo. Current Folder Name: %foldername%
if exist "%foldername%.exe" del /f /q "%foldername%.exe"
SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -ldflags "-s -w" -o %foldername%.exe
:: comment this if you don't have upx or don't want to use upx
if defined UPX_PATH if exist "%UPX_PATH%" "%UPX_PATH%" -3 -v "%foldername%.exe"
