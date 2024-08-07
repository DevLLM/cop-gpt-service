@echo off

where /q npm
if %errorlevel% neq 0 (
    echo Node.js is not installed. Please download it from https://nodejs.org/en/download
    pause
    exit /b
)

where /q github-copilot-cli
if %errorlevel% neq 0 (
    echo github-copilot-cli is not installed. Installing...
    npm i @githubnext/github-copilot-cli -g
    if %errorlevel% neq 0 (
        echo Failed to install github-copilot-cli. Please check your npm installation and try again.
        pause
        exit /b
    )
)

:: Check if token file exists
if exist %USERPROFILE%\.copilot-cli-access-token (
    :: View the token
    type %USERPROFILE%\.copilot-cli-access-token
    echo.

    :: Ask user if they want to reauthorize
    set /p reauth="Token exists. Do you want to reauthorize? (Y/N): "
    if /i "%reauth%" neq "Y" (
        echo Exiting script...
        exit /b
    )
)

:: Retrieve the token
github-copilot-cli auth
if %errorlevel% neq 0 (
    echo Failed to retrieve GitHub Copilot Token. Please check your internet connection and try again.
    pause
    exit /b
)

:: View the token
type %USERPROFILE%\.copilot-cli-access-token

if %errorlevel% neq 0 (
    echo Failed to view the token. Please check the file path and try again.
    pause
    exit /b
)

pause
