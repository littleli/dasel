package dasel

import (
	"fmt"
	"reflect"
)

// propagate recursively propagates the given nodes value up to the root node.
func propagate(n *Node) error {
	if n.Previous == nil {
		return nil
	}

	if !n.propagated || n.wasInitialised {
		if err := propagateValue(n); err != nil {
			return fmt.Errorf("could not propagate value: %w", err)
		}
	}
	return propagate(n.Previous)
}

// propagateValue sends the value of the current node up to the previous node in the chain.
func propagateValue(n *Node) error {
	if n.Previous == nil {
		return nil
	}

	switch n.Selector.Type {
	case "PROPERTY":
		return propagateValueProperty(n)
	case "INDEX":
		return propagateValueIndex(n)
	// case "DYNAMIC":
	// 	return propagateValueDynamic(n)
	case "NEXT_AVAILABLE_INDEX":
		return propagateValueNextAvailableIndex(n)
	default:
		return &UnsupportedSelector{Selector: n.Selector.Type}
	}
}

// propagateValueProperty sends the value of the current node up to the previous node in the chain.
func propagateValueProperty(n *Node) error {
	if !isValid(n.Previous.Value) {
		return &UnexpectedPreviousNilValue{Selector: n.Previous.Selector.Current}
	}

	value := unwrapValue(n.Previous.Value)

	if value.Kind() == reflect.Map {
		value.SetMapIndex(reflect.ValueOf(n.Selector.Property), n.Value)
		// Set propagated to true here since we modified the previous value by reference.
		n.Previous.propagated = true
		return nil
	}

	return &UnsupportedTypeForSelector{Selector: n.Selector, Value: n.Previous.Value}
}

// propagateValueIndex sends the value of the current node up to the previous node in the chain.
func propagateValueIndex(n *Node) error {
	if !isValid(n.Previous.Value) {
		return &UnexpectedPreviousNilValue{Selector: n.Previous.Selector.Current}
	}

	value := unwrapValue(n.Previous.Value)

	if value.Kind() == reflect.Slice {
		if n.Selector.Index >= 0 && n.Selector.Index < value.Len() {
			value.Index(n.Selector.Index).Set(n.Value)
			return nil
		}
		n.Previous.Value = reflect.Append(value, n.Value)
		return nil
	}

	return &UnsupportedTypeForSelector{Selector: n.Selector, Value: value.Kind()}
}

// propagateValueNextAvailableIndex sends the value of the current node up to the previous node in the chain.
func propagateValueNextAvailableIndex(n *Node) error {
	if !isValid(n.Previous.Value) {
		return &UnexpectedPreviousNilValue{Selector: n.Previous.Selector.Current}
	}

	value := unwrapValue(n.Previous.Value)

	if value.Kind() == reflect.Slice {
		n.Previous.Value = reflect.Append(value, n.Value)
		return nil
	}

	return &UnsupportedTypeForSelector{Selector: n.Selector, Value: value.Kind()}
}

// propagateValueDynamic sends the value of the current node up to the previous node in the chain.
func propagateValueDynamic(n *Node) error {
	if !isValid(n.Previous.Value) {
		return &UnexpectedPreviousNilValue{Selector: n.Previous.Selector.Current}
	}

	value := unwrapValue(n.Previous.Value)

	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			object := value.Index(i)
			found, err := processFindDynamicItem(n, object)
			if err != nil {
				return err
			}
			if found {
				object.Set(n.Value)
				return nil
			}
		}
		return &ValueNotFound{Selector: n.Selector.Current, Node: n}
	}

	return &UnsupportedTypeForSelector{Selector: n.Selector, Value: value.Kind()}
}