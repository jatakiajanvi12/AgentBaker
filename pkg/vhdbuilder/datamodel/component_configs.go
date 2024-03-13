package datamodel

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Components is the data model for the component config.
// The component config is a JSON file, for example, vhdbuilder/packer/components.json.
// for example, vhdbuilder/packer/components.json.
type Components struct {
	ContainerImages []*ContainerImage `json:"ContainerImages"`
	DownloadFiles   []DownloadFile    `json:"DownloadFiles"`
}

type ContainerImage struct {
	DownloadURL       string   `json:"downloadURL"`
	Amd64OnlyVersions []string `json:"amd64OnlyVersions"`
	MultiArchVersions []string `json:"multiArchVersions"`
}

type DownloadFile struct {
	FileName               string   `json:"fileName"`
	DownloadLocation       string   `json:"downloadLocation"`
	DownloadURL            string   `json:"downloadURL"`
	Versions               []string `json:"versions"`
	TargetContainerRuntime string   `json:"targetContainerRuntime,omitempty"`
}

// KubeProxyImages is the data model for the kube-proxy image config.
// The kube-proxy image config is a JSON file, for example, vhdbuilder/packer/kube-proxy-images.json.
type KubeProxyImages struct {
	DockerKubeProxyImages     *DockerKubeProxyImages `json:"dockerKubeProxyImages"`
	ContainerdKubeProxyImages *DockerKubeProxyImages `json:"containerdKubeProxyImages"`
}

type DockerKubeProxyImages struct {
	ContainerImages []*ContainerImage `json:"ContainerImages"`
}

func loadJSONFromFile(path string, v interface{}) error {
	configFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	return jsonParser.Decode(&v)
}

func toImageList(downloadURL string, imageTagList []string) ([]string, error) {
	ret := []string{}

	if !strings.Contains(downloadURL, "*") {
		return ret, fmt.Errorf("downloadURL does not contain *")
	}

	for _, tag := range imageTagList {
		ret = append(ret, strings.Replace(downloadURL, "*", tag, 1))
	}

	return ret, nil
}

// begins Components

// NewComponentsFromFile loads component config from the given file and returns a Components object.
func NewComponentsFromFile(path string) (*Components, error) {
	ret := &Components{}

	err := loadJSONFromFile(path, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

/*
ToImageList returns a list of image names from the given Components object.
The image names are generated from the given Components object and the given downloadURL.
The image names are generated by replacing the * in the downloadURL with the image tag.
*/
func (c *Components) ToImageList() []string {
	ret := []string{}

	if c.ContainerImages != nil {
		for _, image := range c.ContainerImages {
			if image.Amd64OnlyVersions != nil {
				amd64OnlyImageList, _ := toImageList(image.DownloadURL, image.Amd64OnlyVersions)
				ret = append(ret, amd64OnlyImageList...)
			}

			if image.MultiArchVersions != nil {
				multiArchImageList, _ := toImageList(image.DownloadURL, image.MultiArchVersions)
				ret = append(ret, multiArchImageList...)
			}
		}
	}
	return ret
}

// ends Components

// begins KubeProxyImages

/*
NewKubeProxyImagesFromFile loads kube-proxy image config from the given file and returns a KubeProxyImages object.
The given file should be a KubeProxyImages object, and should be in JSON format.
The given file should be in JSON format.
*/
func NewKubeProxyImagesFromFile(path string) (*KubeProxyImages, error) {
	ret := &KubeProxyImages{}

	err := loadJSONFromFile(path, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// processProxyImages is a helper function for ToImageList().
func processProxyImages(image *ContainerImage, ret *[]string) error {
	var err error
	var amd64OnlyImageList []string
	var multiArchImageList []string

	if image.Amd64OnlyVersions != nil {
		amd64OnlyImageList, err = toImageList(image.DownloadURL, image.Amd64OnlyVersions)
		if err != nil {
			return err
		}
		*ret = append(*ret, amd64OnlyImageList...)
	}

	if image.MultiArchVersions != nil {
		multiArchImageList, err = toImageList(image.DownloadURL, image.MultiArchVersions)
		if err != nil {
			return err
		}
		*ret = append(*ret, multiArchImageList...)
	}
	return err
}

/*
ToImageList returns a list of image names from the given KubeProxyImages object.
The image names are generated from the given KubeProxyImages object and the given downloadURL.
The image names are generated by replacing the * in the downloadURL with the image tag.
*/
func (k *KubeProxyImages) ToImageList() ([]string, error) {
	var err error
	var ret []string

	if k.DockerKubeProxyImages != nil && k.DockerKubeProxyImages.ContainerImages != nil {
		for _, image := range k.DockerKubeProxyImages.ContainerImages {
			err = processProxyImages(image, &ret)
			if err != nil {
				return ret, err
			}
		}
	}

	if k.ContainerdKubeProxyImages != nil && k.ContainerdKubeProxyImages.ContainerImages != nil {
		for _, image := range k.ContainerdKubeProxyImages.ContainerImages {
			err = processProxyImages(image, &ret)
			if err != nil {
				return ret, err
			}
		}
	}

	return ret, err
}