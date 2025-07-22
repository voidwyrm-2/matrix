buildfor() {
    echo "building $1/$2"
    GOOS="$1" GOARCH="$2" go build -o "$3" .
    echo "built $1/$2"
    zip "build/$1-$2.zip" $3
    rm $3
}

EXE_NAME_UNIX="matrix"
EXE_NAME_WIN="matrix.exe"


rm -rf "$EXE_NAME_UNIX" "$EXE_NAME_WIN"
rm -rf build
mkdir build


buildfor darwin arm64 "$EXE_NAME_UNIX"
buildfor windows arm64 "$EXE_NAME_WIN"
buildfor windows amd64 "$EXE_NAME_WIN"
buildfor linux arm64 "$EXE_NAME_UNIX"
buildfor linux amd64 "$EXE_NAME_UNIX"
