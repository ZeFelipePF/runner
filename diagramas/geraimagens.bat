@echo off
REM Renderiza os diagramas C4 (.puml) em SVG/PNG no Windows.
REM Requer: plantuml no PATH ou Docker Desktop.

setlocal
set DIR=%~dp0
set OUT=%DIR%imagens
if not exist "%OUT%" mkdir "%OUT%"

where plantuml >nul 2>nul
if %ERRORLEVEL%==0 (
    echo Usando plantuml local...
    plantuml -tsvg -o "%OUT%" "%DIR%\*.puml"
    plantuml -tpng -o "%OUT%" "%DIR%\*.puml"
    goto fim
)

where docker >nul 2>nul
if %ERRORLEVEL%==0 (
    echo Usando plantuml via Docker...
    docker run --rm -v "%DIR%":/data plantuml/plantuml -tsvg -o /data/imagens /data/*.puml
    docker run --rm -v "%DIR%":/data plantuml/plantuml -tpng -o /data/imagens /data/*.puml
    goto fim
)

echo ERRO: instale plantuml ou Docker Desktop. >&2
exit /b 1

:fim
echo OK. SVGs/PNGs em %OUT%
