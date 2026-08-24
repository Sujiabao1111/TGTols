
:: get folder name without path
for /f "delims=" %%A in ('cd') do (
    set foldername=%%~nxA
)

echo. Current Folder Name: %foldername%
rm %foldername%.exe
SET CGO_ENABLED=0
SET GOOS=windows
SET GOARCH=amd64
go build -ldflags "-s -w" -o %foldername%.exe
:: comment this if you don't have upx or don't want to use upx
E:\upx-3.96-win64\upx.exe -3 -v %foldername%.exe