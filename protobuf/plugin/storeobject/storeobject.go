package storeobject

import (
	"github.com/docker/swarmkit/protobuf/plugin"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// FIXME(aaronl): Look at fields inside the descriptor instead of
// special-casing based on name.
var typesWithNoSpec = map[string]struct{}{
	"Task":      {},
	"Resource":  {},
	"Extension": {},
}

type storeObjectGen struct {
	*generator.Generator
	generator.PluginImports
	eventsPkg  generator.Single
	stringsPkg generator.Single
}

func init() {
	generator.RegisterPlugin(new(storeObjectGen))
}

func (d *storeObjectGen) Name() string {
	return "storeobject"
}

func (d *storeObjectGen) Init(g *generator.Generator) {
	d.Generator = g
}

func (d *storeObjectGen) genMsgStoreObject(m *generator.Descriptor, storeObject *plugin.StoreObject) {
	ccTypeName := generator.CamelCaseSlice(m.TypeName())

	// Generate event types

	d.P("type ", ccTypeName, "CheckFunc func(t1, t2 *", ccTypeName, ") bool")
	d.P()

	for _, event := range []string{"Create", "Update", "Delete"} {
		d.P("type Event", event, ccTypeName, " struct {")
		d.In()
		d.P(ccTypeName, " *", ccTypeName)
		if event == "Update" {
			d.P("Old", ccTypeName, " *", ccTypeName)
		}
		d.P("Checks []", ccTypeName, "CheckFunc")
		d.Out()
		d.P("}")
		d.P()
		d.P("func (e Event", event, ccTypeName, ") Matches(apiEvent ", d.eventsPkg.Use(), ".Event) bool {")
		d.In()
		d.P("typedEvent, ok := apiEvent.(Event", event, ccTypeName, ")")
		d.P("if !ok {")
		d.In()
		d.P("return false")
		d.Out()
		d.P("}")
		d.P()
		d.P("for _, check := range e.Checks {")
		d.In()
		d.P("if !check(e.", ccTypeName, ", typedEvent.", ccTypeName, ") {")
		d.In()
		d.P("return false")
		d.Out()
		d.P("}")
		d.Out()
		d.P("}")
		d.P("return true")
		d.Out()
		d.P("}")
	}

	// Generate methods for this type

	d.P("func (m *", ccTypeName, ") CopyStoreObject() StoreObject {")
	d.In()
	d.P("return m.Copy()")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") GetMeta() Meta {")
	d.In()
	d.P("return m.Meta")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") SetMeta(meta Meta) {")
	d.In()
	d.P("m.Meta = meta")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") GetID() string {")
	d.In()
	d.P("return m.ID")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") EventCreate() Event {")
	d.In()
	d.P("return EventCreate", ccTypeName, "{", ccTypeName, ": m}")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") EventUpdate(oldObject StoreObject) Event {")
	d.In()
	d.P("if oldObject != nil {")
	d.In()
	d.P("return EventUpdate", ccTypeName, "{", ccTypeName, ": m, Old", ccTypeName, ": oldObject.(*", ccTypeName, ")}")
	d.Out()
	d.P("} else {")
	d.In()
	d.P("return EventUpdate", ccTypeName, "{", ccTypeName, ": m}")
	d.Out()
	d.P("}")
	d.Out()
	d.P("}")
	d.P()

	d.P("func (m *", ccTypeName, ") EventDelete() Event {")
	d.In()
	d.P("return EventDelete", ccTypeName, "{", ccTypeName, ": m}")
	d.Out()
	d.P("}")
	d.P()

	// Generate event check functions

	if storeObject.WatchSelectors.ID != nil && *storeObject.WatchSelectors.ID {
		d.P("func ", ccTypeName, "CheckID(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.ID == v2.ID")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.IDPrefix != nil && *storeObject.WatchSelectors.IDPrefix {
		d.P("func ", ccTypeName, "CheckIDPrefix(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return ", d.stringsPkg.Use(), ".HasPrefix(v2.ID, v1.ID)")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.Name != nil && *storeObject.WatchSelectors.Name {
		d.P("func ", ccTypeName, "CheckName(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		// Node is a special case
		if *m.Name == "Node" {
			d.P("if v1.Description == nil || v2.Description == nil {")
			d.In()
			d.P("return false")
			d.Out()
			d.P("}")
			d.P("return v1.Description.Hostname == v2.Description.Hostname")
		} else if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
			d.P("return v1.Annotations.Name == v2.Annotations.Name")
		} else {
			d.P("return v1.Spec.Annotations.Name == v2.Spec.Annotations.Name")
		}
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.NamePrefix != nil && *storeObject.WatchSelectors.NamePrefix {
		d.P("func ", ccTypeName, "CheckNamePrefix(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		// Node is a special case
		if *m.Name == "Node" {
			d.P("if v1.Description == nil || v2.Description == nil {")
			d.In()
			d.P("return false")
			d.Out()
			d.P("}")
			d.P("return ", d.stringsPkg.Use(), ".HasPrefix(v2.Description.Hostname, v1.Description.Hostname)")
		} else if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
			d.P("return ", d.stringsPkg.Use(), ".HasPrefix(v2.Annotations.Name, v1.Annotations.Name)")
		} else {
			d.P("return ", d.stringsPkg.Use(), ".HasPrefix(v2.Spec.Annotations.Name, v1.Spec.Annotations.Name)")
		}
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.Custom != nil && *storeObject.WatchSelectors.Custom {
		d.P("func ", ccTypeName, "CheckCustom(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		// Node is a special case
		if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
			d.P("return checkCustom(v1.Annotations, v2.Annotations)")
		} else {
			d.P("return checkCustom(v1.Spec.Annotations, v2.Spec.Annotations)")
		}
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.CustomPrefix != nil && *storeObject.WatchSelectors.CustomPrefix {
		d.P("func ", ccTypeName, "CheckCustomPrefix(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		// Node is a special case
		if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
			d.P("return checkCustomPrefix(v1.Annotations, v2.Annotations)")
		} else {
			d.P("return checkCustomPrefix(v1.Spec.Annotations, v2.Spec.Annotations)")
		}
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.NodeID != nil && *storeObject.WatchSelectors.NodeID {
		d.P("func ", ccTypeName, "CheckNodeID(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.NodeID == v2.NodeID")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.ServiceID != nil && *storeObject.WatchSelectors.ServiceID {
		d.P("func ", ccTypeName, "CheckServiceID(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.ServiceID == v2.ServiceID")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.Slot != nil && *storeObject.WatchSelectors.Slot {
		d.P("func ", ccTypeName, "CheckSlot(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.Slot == v2.Slot")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.DesiredState != nil && *storeObject.WatchSelectors.DesiredState {
		d.P("func ", ccTypeName, "CheckDesiredState(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.DesiredState == v2.DesiredState")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.Role != nil && *storeObject.WatchSelectors.Role {
		d.P("func ", ccTypeName, "CheckRole(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.Role == v2.Role")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.Membership != nil && *storeObject.WatchSelectors.Membership {
		d.P("func ", ccTypeName, "CheckMembership(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.Spec.Membership == v2.Spec.Membership")
		d.Out()
		d.P("}")
		d.P()
	}

	if storeObject.WatchSelectors.Kind != nil && *storeObject.WatchSelectors.Kind {
		d.P("func ", ccTypeName, "CheckKind(v1, v2 *", ccTypeName, ") bool {")
		d.In()
		d.P("return v1.Kind == v2.Kind")
		d.Out()
		d.P("}")
		d.P()
	}

	// Generate indexer by ID

	d.P("type ", ccTypeName, "IndexerByID struct{}")
	d.P()

	d.genFromArgs(ccTypeName + "IndexerByID")
	d.genPrefixFromArgs(ccTypeName + "IndexerByID")

	d.P("func (indexer ", ccTypeName, "IndexerByID) FromObject(obj interface{}) (bool, []byte, error) {")
	d.In()
	d.P("m := obj.(*", ccTypeName, ")")
	// Add the null character as a terminator
	d.P(`return true, []byte(m.ID + "\x00"), nil`)
	d.Out()
	d.P("}")

	// Generate indexer by name

	d.P("type ", ccTypeName, "IndexerByName struct{}")
	d.P()

	d.genFromArgs(ccTypeName + "IndexerByName")
	d.genPrefixFromArgs(ccTypeName + "IndexerByName")

	d.P("func (indexer ", ccTypeName, "IndexerByName) FromObject(obj interface{}) (bool, []byte, error) {")
	d.In()
	d.P("m := obj.(*", ccTypeName, ")")
	if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
		d.P(`val := m.Annotations.Name`)
	} else {
		d.P(`val := m.Spec.Annotations.Name`)
	}
	// Add the null character as a terminator
	d.P("return true, []byte(", d.stringsPkg.Use(), `.ToLower(val) + "\x00"), nil`)
	d.Out()
	d.P("}")

	// Generate custom indexer

	d.P("type ", ccTypeName, "CustomIndexer struct{}")
	d.P()

	d.genFromArgs(ccTypeName + "CustomIndexer")
	d.genPrefixFromArgs(ccTypeName + "CustomIndexer")

	d.P("func (indexer ", ccTypeName, "CustomIndexer) FromObject(obj interface{}) (bool, [][]byte, error) {")
	d.In()
	d.P("m := obj.(*", ccTypeName, ")")
	if _, hasNoSpec := typesWithNoSpec[*m.Name]; hasNoSpec {
		d.P(`return customIndexer("", &m.Annotations)`)
	} else {
		d.P(`return customIndexer("", &m.Spec.Annotations)`)
	}
	d.Out()
	d.P("}")
}

func (d *storeObjectGen) genFromArgs(indexerName string) {
	d.P("func (indexer ", indexerName, ") FromArgs(args ...interface{}) ([]byte, error) {")
	d.In()
	d.P("return fromArgs(args...)")
	d.Out()
	d.P("}")
}

func (d *storeObjectGen) genPrefixFromArgs(indexerName string) {
	d.P("func (indexer ", indexerName, ") PrefixFromArgs(args ...interface{}) ([]byte, error) {")
	d.In()
	d.P("return prefixFromArgs(args...)")
	d.Out()
	d.P("}")

}

func (d *storeObjectGen) genNewStoreAction(topLevelObjs []string) {
	if len(topLevelObjs) == 0 {
		return
	}

	// Generate NewStoreAction
	d.P("func NewStoreAction(c Event) (StoreAction, error) {")
	d.In()
	d.P("var sa StoreAction")
	d.P("switch v := c.(type) {")
	for _, ccTypeName := range topLevelObjs {
		d.P("case EventCreate", ccTypeName, ":")
		d.In()
		d.P("sa.Action = StoreActionKindCreate")
		d.P("sa.Target = &StoreAction_", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}")
		d.Out()
		d.P("case EventUpdate", ccTypeName, ":")
		d.In()
		d.P("sa.Action = StoreActionKindUpdate")
		d.P("sa.Target = &StoreAction_", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}")
		d.Out()
		d.P("case EventDelete", ccTypeName, ":")
		d.In()
		d.P("sa.Action = StoreActionKindRemove")
		d.P("sa.Target = &StoreAction_", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}")
		d.Out()
	}
	d.P("default:")
	d.In()
	d.P("return StoreAction{}, errUnknownStoreAction")
	d.Out()
	d.P("}")
	d.P("return sa, nil")
	d.Out()
	d.P("}")
	d.P()
}

func (d *storeObjectGen) genWatchMessageEvent(topLevelObjs []string) {
	if len(topLevelObjs) == 0 {
		return
	}

	// Generate WatchMessageEvent
	d.P("func WatchMessageEvent(c Event) *WatchMessage_Event {")
	d.In()
	d.P("switch v := c.(type) {")
	for _, ccTypeName := range topLevelObjs {
		d.P("case EventCreate", ccTypeName, ":")
		d.In()
		d.P("return &WatchMessage_Event{Action: WatchActionKindCreate, Object: &Object{Object: &Object_", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}}}")
		d.Out()
		d.P("case EventUpdate", ccTypeName, ":")
		d.In()
		d.P("if v.Old", ccTypeName, " != nil {")
		d.In()
		d.P("return &WatchMessage_Event{Action: WatchActionKindUpdate, Object: &Object{Object: &Object_", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}}, OldObject: &Object{Object: &Object_", ccTypeName, "{", ccTypeName, ": v.Old", ccTypeName, "}}}")
		d.Out()
		d.P("} else {")
		d.In()
		d.P("return &WatchMessage_Event{Action: WatchActionKindUpdate, Object: &Object{Object: &Object_", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}}}")
		d.Out()
		d.P("}")
		d.Out()
		d.P("case EventDelete", ccTypeName, ":")
		d.In()
		d.P("return &WatchMessage_Event{Action: WatchActionKindRemove, Object: &Object{Object: &Object_", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}}}")
		d.Out()
	}
	d.P("}")
	d.P("return nil")
	d.Out()
	d.P("}")
	d.P()
}

func (d *storeObjectGen) genEventFromStoreAction(topLevelObjs []string) {
	if len(topLevelObjs) == 0 {
		return
	}

	// Generate EventFromStoreAction
	d.P("func EventFromStoreAction(sa StoreAction, oldObject StoreObject) (Event, error) {")
	d.In()
	d.P("switch v := sa.Target.(type) {")
	for _, ccTypeName := range topLevelObjs {
		d.P("case *StoreAction_", ccTypeName, ":")
		d.In()
		d.P("switch sa.Action {")

		d.P("case StoreActionKindCreate:")
		d.In()
		d.P("return EventCreate", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}, nil")
		d.Out()

		d.P("case StoreActionKindUpdate:")
		d.In()
		d.P("if oldObject != nil {")
		d.In()
		d.P("return EventUpdate", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, ", Old", ccTypeName, ": oldObject.(*", ccTypeName, ")}, nil")
		d.Out()
		d.P("} else {")
		d.In()
		d.P("return EventUpdate", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}, nil")
		d.Out()
		d.P("}")
		d.Out()

		d.P("case StoreActionKindRemove:")
		d.In()
		d.P("return EventDelete", ccTypeName, "{", ccTypeName, ": v.", ccTypeName, "}, nil")
		d.Out()

		d.P("}")
		d.Out()
	}
	d.P("}")
	d.P("return nil, errUnknownStoreAction")
	d.Out()
	d.P("}")
	d.P()
}

func (d *storeObjectGen) Generate(file *generator.FileDescriptor) {
	d.PluginImports = generator.NewPluginImports(d.Generator)
	d.eventsPkg = d.NewImport("github.com/docker/go-events")
	d.stringsPkg = d.NewImport("strings")

	var topLevelObjs []string

	for _, m := range file.Messages() {
		if m.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}

		if m.Options == nil {
			continue
		}
		storeObjIntf, err := proto.GetExtension(m.Options, plugin.E_StoreObject)
		if err != nil {
			// no StoreObject extension
			continue
		}

		d.genMsgStoreObject(m, storeObjIntf.(*plugin.StoreObject))

		topLevelObjs = append(topLevelObjs, generator.CamelCaseSlice(m.TypeName()))
	}

	d.genNewStoreAction(topLevelObjs)
	d.genEventFromStoreAction(topLevelObjs)
	d.genWatchMessageEvent(topLevelObjs)
}
