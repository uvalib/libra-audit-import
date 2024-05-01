//
//
//

package main

import (
	"fmt"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
	"regexp"
	"strings"
)

var etdDeletedMsg = "Deleted"
var etdFileField = "File"

// Libra ETD event exports take the form:
// timestamp \t who \t work identifier \t message

const (
	etdTimeIx   = 0
	etdWhoIx    = 1
	etdWorkIdIx = 2
	etdMsgIx    = 3
)

func makeEtdEvent(namespace string, auditLine string) (*uvalibrabus.UvaBusEvent, error) {

	parts := strings.Split(auditLine, "\t")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid audit line: %s", auditLine)
	}

	ignore, field, before, after := parseEtdAuditMessage(parts[etdMsgIx])
	if ignore == true {
		return nil, nil
	}

	auditEvent := uvalibrabus.UvaAuditEvent{
		Who:       parts[etdWhoIx],
		FieldName: field,
		Before:    before,
		After:     after,
	}
	buf, err := auditEvent.Serialize()
	if err != nil {
		return nil, err
	}

	logDebug(fmt.Sprintf("audit event: %s", auditEvent.String()))

	busEvent := uvalibrabus.UvaBusEvent{
		EventName:  uvalibrabus.EventFieldUpdate,
		Namespace:  namespace,
		Identifier: parts[etdWorkIdIx],
		EventTime:  fixTimeStamp(parts[etdTimeIx]),
		Detail:     buf,
	}
	return &busEvent, nil
}

func parseEtdAuditMessage(msg string) (bool, string, string, string) {

	if msg == etdDeletedMsg {
		//fmt.Printf("IGNORING [%s]\n", msg)
		return true, "", "", ""
	}

	// is this a file add?
	fileAddRegex := `^File (.*) added$`
	re := regexp.MustCompile(fileAddRegex)
	matches := re.FindStringSubmatch(msg)
	if matches != nil && len(matches) == 2 {
		//fmt.Printf("FILE ADD [%s]\n", matches[1])
		return false, fileFieldName, "", matches[1]
	}

	// is this a file delete?
	fileDelRegex := `^File (.*) deleted$`
	re = regexp.MustCompile(fileDelRegex)
	matches = re.FindStringSubmatch(msg)
	if matches != nil && len(matches) == 2 {
		//fmt.Printf("FILE DEL [%s]\n", matches[1])
		return false, fileFieldName, matches[1], ""
	}

	// is this a field update?
	fieldUpdateRegex := `^(.*) updated from: '(.*)' to: '(.*)'$`
	re = regexp.MustCompile(fieldUpdateRegex)
	matches = re.FindStringSubmatch(msg)
	if matches != nil && len(matches) == 4 {
		//fmt.Printf("FIELD [%s] [%s] -> [%s]\n", matches[1], matches[2], matches[3])
		return false, matches[1], matches[2], matches[3]
	}

	// special case of Admin Notes being updated
	adminNotesUpdateRegex := `^Admin Notes updated to include '(.*)'$`
	re = regexp.MustCompile(adminNotesUpdateRegex)
	matches = re.FindStringSubmatch(msg)
	if matches != nil && len(matches) == 2 {
		//fmt.Printf("ADMIN NOTES [%s]\n", matches[1])
		return false, "Admin Notes", "append", matches[1]
	}

	// cant parse this so ignore it
	//fmt.Printf("UNSURE [%s]\n", msg)
	//fmt.Printf("IGNORING [%s]\n", msg)
	return true, "", "", ""
}

//
// end of file
//
