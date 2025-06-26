/*
Other packages expect that this file provides
3 functions which should behave like
bigtable, err := NewBigTableWriter(ctx, ProjectId, InstanceName, TableName)
_ = bigtable.WriteMessage(ctx, "ai", resp)
messages, err := bigtable.ReadMessages(ctx)
*/

package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"strconv"

	"cloud.google.com/go/bigtable"
)

// TODO: rename this to something without the word "writer" e.g. CustomBigtableStruct
type BigtableWriter struct {
	table *bigtable.Table
}


// TODO: [understand] figure out what all context.Context contains
func NewBigTableWriter(ctx context.Context, projectId string, instanceId string, tableName string) (*BigtableWriter, error) {
	client, err := bigtable.NewClient(ctx, projectId, instanceId)
	if err != nil {
		log.Fatal("Could not create bigtable client", err)	
	}

	table := client.Open(tableName)
	bigtableWriter := BigtableWriter{ table: table }
	return &bigtableWriter, nil
}

const ColumnFamilyName = "chats"
const SenderColumn = "sender"
const MessageColumn = "message"

// senderType is one of "user" or "ai"
// TODO: make an enum type thing out of senderType to run validation against.
func (writer *BigtableWriter) WriteMessage(ctx context.Context, senderType string, message string) error {
	
	unixTimestamp := time.Now().UnixNano()

	// TODO (understand) what Sprintf does
	rowKey := fmt.Sprintf("chat-%d", unixTimestamp)

	mut := bigtable.NewMutation()
	mut.Set(ColumnFamilyName, SenderColumn, bigtable.Now(), []byte(senderType))
	mut.Set(ColumnFamilyName, MessageColumn, bigtable.Now(), []byte(message))

	err := writer.table.Apply(ctx, rowKey, mut)
	if err != nil { return err }

	log.Printf("Saved chat message to Bigtable: %s -> %s", senderType, message)
	return nil
}


type ChatMessage struct {
	Timestamp int64 `json:"timestamp"`
	Sender string `json:"sender"`
	Message string `json:"message"`
}

func (b *BigtableWriter) ReadMessages(ctx context.Context) ([]ChatMessage, error) {
	var messages []ChatMessage

	err := b.table.ReadRows(ctx, bigtable.PrefixRange("chat-"), func(row bigtable.Row) bool {
		msg := ChatMessage{}

		// log.Printf("DEBUG: row key = %s, row contents = %+v", row.Key(), row)
		
		
		// TODO: understand what this 5: does
		// exampel of row.Key() is chat-17509048129787360002025/06/25 22:39:17
		if ts, err := strconv.ParseInt(row.Key()[5:], 10, 64); err == nil {
			msg.Timestamp = ts
		}
		
		// row maps string -> arr of ReadItem
		if readItems := row[ColumnFamilyName]; len(readItems) > 0 {
			for _, item := range readItems {
				if item.Column == ColumnFamilyName + ":" + SenderColumn  {
					msg.Sender = string(item.Value)
				} 
				if item.Column == ColumnFamilyName + ":" + MessageColumn {
					msg.Message = string(item.Value)
				}
			}
		}

		messages = append(messages, msg)
		return true
	})

	if err != nil {
		return nil, err
	}
	
	return messages, nil
}