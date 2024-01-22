package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const ErrorCode = 12

func formatBodyTypeError(body map[string]any, errorMessage string) map[string]any {
	body["type"] = "error"
	body["code"] = ErrorCode
	body["text"] = errorMessage
	delete(body, "topology")
	return body
}

func addMessageIfNotExists(slice []any, item any) ([]any, bool) {
	// Check if the item exists in the slice
	for _, existingItem := range slice {
		if existingItem == item {
			return slice, false // item already exists, return the original slice and false
		}
	}

	// Item doesn't exist, so append it to the slice
	updatedSlice := append(slice, item)
	return updatedSlice, true
}

func main() {
	n := maelstrom.NewNode()

	var neighborNodes []any

	n.Handle("topology", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Extract and store neighbor nodes from the received topology information.
		if topology, ok := body["topology"].(map[string]any); ok {
			if neighbors, ok := topology[n.ID()].([]any); ok {
				neighborNodes = append(neighborNodes, neighbors...)
			} else {
				body = formatBodyTypeError(body, "The node is not a string")
				return n.Reply(msg, body)
			}
		} else {
			body = formatBodyTypeError(body, "The topology node is not a any")
			return n.Reply(msg, body)
		}

		delete(body, "topology")
		body["type"] = "topology_ok"
		return n.Reply(msg, body)
	})

	var messages []any

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		var shouldSendMessage bool
		messages, shouldSendMessage = addMessageIfNotExists(messages, body["message"])
		if shouldSendMessage {
			for _, node := range neighborNodes {
				// Ensure we don't send the message back to the sender.
				if node, ok := node.(string); ok && node != msg.Src {
					n.Send(node, body)
				}
			}
		}

		delete(body, "message")
		body["type"] = "broadcast_ok"
		return n.Reply(msg, body)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "read_ok"
		body["messages"] = messages
		return n.Reply(msg, body)
	})

	n.Handle("broadcast_ok", func(msg maelstrom.Message) error {
		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
