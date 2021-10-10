# Versions used:
#	protoc:        v3.17.3
#	protoc-gen-go: v1.25.0-devel

for X in $(find . -name "*.proto" | sed "s|^\./||"); do
	protoc -I$(pwd) --go_out=paths=source_relative:. $X
done