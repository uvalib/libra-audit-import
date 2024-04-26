//
//
//

package main

import (
	"fmt"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
	"strings"
)

// Libra Open event exports take the form:
// timestamp \t who \t work identifier \t field \t before \t after

const (
	openTimeIx   = 0
	openWhoIx    = 1
	openWorkIdIx = 2
	openFieldIx  = 3
	openBeforeIx = 4
	openAfterIx  = 5
)

func makeOpenEvent(namespace string, auditLine string) (*uvalibrabus.UvaBusEvent, error) {

	parts := strings.Split(auditLine, "\t")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid audit line: %s", auditLine)
	}

	var buf []byte
	var err error
	//var eventName string

	// audit events that refer to a 'file' are handled differently
	//	if parts[openFieldIx] == fileField {
	//
	//		before := ""
	//		after := ""
	//
	//		if parts[openBeforeIx] == emptyStr {
	//			eventName = uvalibrabus.EventFileCreate
	//		} else if parts[openAfterIx] == emptyStr {
	//			eventName = uvalibrabus.EventFileDelete
	//		} else {
	//			eventName = uvalibrabus.EventFileUpdate
	//		}
	//
	//	} else {
	//eventName = uvalibrabus.EventFieldUpdate
	auditEvent := uvalibrabus.UvaAuditEvent{
		Who:       parts[openWhoIx],
		FieldName: parts[openFieldIx],
		Before:    parts[openBeforeIx],
		After:     parts[openAfterIx],
	}
	buf, err = auditEvent.Serialize()
	if err != nil {
		return nil, err
	}
	//	}

	busEvent := uvalibrabus.UvaBusEvent{
		EventName:  uvalibrabus.EventFieldUpdate,
		Namespace:  namespace,
		Identifier: parts[openWorkIdIx],
		EventTime:  fixTimeStamp(parts[openTimeIx]),
		Detail:     buf,
	}
	return &busEvent, nil
}

//
// end of file
//
