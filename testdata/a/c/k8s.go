package c

import (
	"fmt"
)

type Object interface {
	DeepCopyObject() Object
}

// KubePkg
// +gengo:deepcopy
// +gengo:deepcopy:interfaces=Object
type KubePkg struct {
	// metav1.TypeMeta   `json:",inline"`
	// metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   KubePkgSpec   `json:"spec,omitempty"`
	Status KubePkgStatus `json:"status,omitempty"`
}

// +gengo:deepcopy
type KubePkgSpec struct {
	// app 版本
	Version string `json:"version"`
	// images 列表
	Images map[string]string `json:"images,omitempty"`
	// k8s manifests
	// 可嵌套
	Manifests Manifests `json:"manifests,omitempty"`
}

type Manifests = map[string]any

// +gengo:deepcopy
type KubePkgStatus struct {
	Statuses Statuses     `json:"statuses,omitempty"`
	Digests  []DigestMeta `json:"digests,omitempty"`
}

type Statuses = map[string]any

// +gengo:deepcopy
type DigestMeta struct {
	Type     string   `json:"type"`
	Digest   string   `json:"digest"`
	Name     string   `json:"name"`
	Size     FileSize `json:"size"`
	Tag      string   `json:"tag,omitempty"`
	Platform string   `json:"platform,omitempty"`
}

// +gengo:deepcopy
type FileSize int64

func (f FileSize) String() string {
	b := int64(f)
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
