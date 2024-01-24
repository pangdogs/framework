//go:generate go run k8s.io/code-generator/cmd/deepcopy-gen -h=./boilerplate.txt -i=./plugins/discovery -o=. -O=generated_deepcopy
package framework
