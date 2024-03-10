package utils

// VyosPath represents a chain of configuration levels
type VyosPath struct {
	Path     []string
	NodePath []string
}

// DYNAMIC_NODE represents the node path name that will be replaced by a "value", e.g. firewall names, interfaces, subnets
var DYNAMIC_NODE = "node.tag"

// VyosPathComponent contains the name of a configuration component and whether it is dynamic
type VyosPathComponent struct {
	Name      string
	IsDynamic bool
}

// MakeVyosPath returns a new and empty path
func MakeVyosPath() *VyosPath {
	return &VyosPath{
		Path:     make([]string, 0),
		NodePath: make([]string, 0),
	}
}

// MakeVyosPC returns a new path component, e.g. firewall -> name are both static components
func MakeVyosPC(name string) VyosPathComponent {
	return VyosPathComponent{Name: name, IsDynamic: false}
}

// MakeVyosDynamicPC returns a new dynamic path component, e.g. "something" in firewall -> name -> <something>
func MakeVyosDynamicPC(name string) VyosPathComponent {
	return VyosPathComponent{Name: name, IsDynamic: true}
}

// Append takes any number of components and adds it to this path
func (v *VyosPath) Append(components ...VyosPathComponent) *VyosPath {
	for _, component := range components {
		v.Path = append(v.Path, component.Name)
		if component.IsDynamic {
			v.NodePath = append(v.NodePath, DYNAMIC_NODE)
		} else {
			v.NodePath = append(v.NodePath, component.Name)
		}
	}

	return v
}

// Extend makes a duplicate of this path and extends it by any number of components
func (v *VyosPath) Extend(components ...VyosPathComponent) *VyosPath {
	return v.DivergeFrom(0, components...)
}

// DivergeFrom makes a duplicate of this path, omitting the previous N entries
func (v *VyosPath) DivergeFrom(omitCount int, components ...VyosPathComponent) *VyosPath {
	v2 := &VyosPath{
		Path:     CopySlice(AllExcept(v.Path, omitCount)),
		NodePath: CopySlice(AllExcept(v.NodePath, omitCount)),
	}

	return v2.Append(components...)
}
