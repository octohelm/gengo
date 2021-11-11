/*
Package c GENERATED BY gengo:deepcopy 
DON'T EDIT THIS FILE
*/
package c

import (
	k8s_io_apimachinery_pkg_runtime "k8s.io/apimachinery/pkg/runtime"
)

func (in *KubePkg) DeepCopyObject() k8s_io_apimachinery_pkg_runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *KubePkg) DeepCopy() *KubePkg {
	if in == nil {
		return nil
	}
	out := new(KubePkg)
	in.DeepCopyInto(out)
	return out
}

func (in *KubePkg) DeepCopyInto(out *KubePkg) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)

}
func (in *KubePkgSpec) DeepCopy() *KubePkgSpec {
	if in == nil {
		return nil
	}
	out := new(KubePkgSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *KubePkgSpec) DeepCopyInto(out *KubePkgSpec) {
	out.Version = in.Version
	if in.Images != nil {
		i, o := &in.Images, &out.Images
		*o = make(map[string]string, len(*i))
		for key, val := range *i {
			(*o)[key] = val
		}
	}
	out.Manifests = in.Manifests.DeepCopy()

}
func (in Manifests) DeepCopy() Manifests {
	if in == nil {
		return nil
	}
	out := make(Manifests)
	in.DeepCopyInto(out)
	return out
}

func (in Manifests) DeepCopyInto(out Manifests) {
	for k := range in {
		out[k] = in[k]
	}
}

func (in *KubePkgStatus) DeepCopy() *KubePkgStatus {
	if in == nil {
		return nil
	}
	out := new(KubePkgStatus)
	in.DeepCopyInto(out)
	return out
}

func (in *KubePkgStatus) DeepCopyInto(out *KubePkgStatus) {
	out.Statuses = in.Statuses.DeepCopy()
	if in.Digests != nil {
		i, o := &in.Digests, &out.Digests
		*o = make([]DigestMeta, len(*i))
		copy(*o, *i)
	}

}
func (in Statuses) DeepCopy() Statuses {
	if in == nil {
		return nil
	}
	out := make(Statuses)
	in.DeepCopyInto(out)
	return out
}

func (in Statuses) DeepCopyInto(out Statuses) {
	for k := range in {
		out[k] = in[k]
	}
}
