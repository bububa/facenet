//go:build cv4
// +build cv4

package main

import (
	"github.com/bububa/camera"
	"github.com/bububa/camera/cv4"
)

func getDevice(opts camera.Options) (camera.Device, error) {
	return cv4.New(opts)
}
