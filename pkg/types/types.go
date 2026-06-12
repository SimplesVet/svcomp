package types

import "strings"

type ObjectKind string

const (
	KindTable     ObjectKind = "table"
	KindView      ObjectKind = "view"
	KindFunction  ObjectKind = "function"
	KindProcedure ObjectKind = "procedure"
	KindTrigger   ObjectKind = "trigger"
	KindEvent     ObjectKind = "event"
)

var OrderedKinds = []ObjectKind{
	KindTable,
	KindView,
	KindFunction,
	KindProcedure,
	KindTrigger,
	KindEvent,
}

type DBObject struct {
	Name       string
	Kind       ObjectKind
	Definition string
}

type DiffAction string

const (
	ActionCreate DiffAction = "create"
	ActionUpdate DiffAction = "update"
	ActionDelete DiffAction = "delete"
)

type DiffResult struct {
	Object DBObject
	Action DiffAction
}

func ObjectKey(kind ObjectKind, name string) string {
	return strings.ToLower(string(kind) + ":" + name)
}

func (o DBObject) Key() string {
	return ObjectKey(o.Kind, o.Name)
}
