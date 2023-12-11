package filestream

import (
	"fmt"

	"github.com/segmentio/encoding/json"

	"github.com/wandb/wandb/core/internal/corelib"
	"github.com/wandb/wandb/core/pkg/service"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var boolTrue bool = true

type processedChunk struct {
	fileType   ChunkTypeEnum
	fileLine   string
	Complete   *bool
	Exitcode   *int32
	Preempting bool
	Uploaded   []string
}

func (fs *FileStream) addProcess(rec *service.Record) {
	fs.processChan <- rec
}

func (fs *FileStream) processRecord(record *service.Record) {
	switch x := record.RecordType.(type) {
	case *service.Record_History:
		fs.streamHistory(x.History)
	case *service.Record_Summary:
		fs.streamSummary(x.Summary)
	case *service.Record_Stats:
		fs.streamSystemMetrics(x.Stats)
	case *service.Record_OutputRaw:
		fs.streamOutputRaw(x.OutputRaw)
	case *service.Record_Exit:
		fs.streamFinish(x.Exit)
	case *service.Record_Preempting:
		fs.streamPreempting(x.Preempting)
	case *service.Record_StreamData:
		fs.streamStreamData(x.StreamData)
	case nil:
		err := fmt.Errorf("filestream: field not set")
		fs.logger.CaptureFatalAndPanic("filestream error:", err)
	default:
		err := fmt.Errorf("filestream: Unknown type %T", x)
		fs.logger.CaptureFatalAndPanic("filestream error:", err)
	}
}

func (fs *FileStream) loopProcess(inChan <-chan protoreflect.ProtoMessage) {
	fs.logger.Debug("filestream: open", "path", fs.path)

	for message := range inChan {
		fs.logger.Debug("filestream: record", "message", message)
		switch x := message.(type) {
		case *service.Record:
			fs.processRecord(x)
		case *service.FilesUploaded:
			fs.streamFilesUploaded(x)
		case nil:
			err := fmt.Errorf("filestream: field not set")
			fs.logger.CaptureFatalAndPanic("filestream error:", err)
		default:
			err := fmt.Errorf("filestream: Unknown type %T", x)
			fs.logger.CaptureFatalAndPanic("filestream error:", err)
		}
	}
}

func (fs *FileStream) streamHistory(msg *service.HistoryRecord) {
	line, err := corelib.JsonifyItems(msg.Item)
	if err != nil {
		fs.logger.CaptureFatalAndPanic("json unmarshal error", err)
	}
	fs.addTransmit(processedChunk{
		fileType: HistoryChunk,
		fileLine: line,
	})
}

func (fs *FileStream) streamSummary(msg *service.SummaryRecord) {
	line, err := corelib.JsonifyItems(msg.Update)
	if err != nil {
		fs.logger.CaptureFatalAndPanic("json unmarshal error", err)
	}
	fs.addTransmit(processedChunk{
		fileType: SummaryChunk,
		fileLine: line,
	})
}

func (fs *FileStream) streamOutputRaw(msg *service.OutputRawRecord) {
	fs.addTransmit(processedChunk{
		fileType: OutputChunk,
		fileLine: msg.Line,
	})
}

func (fs *FileStream) streamSystemMetrics(msg *service.StatsRecord) {
	// todo: there is a lot of unnecessary overhead here,
	//  we should prepare all the data in the system monitor
	//  and then send it in one record
	row := make(map[string]interface{})
	row["_wandb"] = true
	timestamp := float64(msg.GetTimestamp().Seconds) + float64(msg.GetTimestamp().Nanos)/1e9
	row["_timestamp"] = timestamp
	row["_runtime"] = timestamp - fs.settings.XStartTime.GetValue()

	for _, item := range msg.Item {
		var val interface{}
		if err := json.Unmarshal([]byte(item.ValueJson), &val); err != nil {
			e := fmt.Errorf("json unmarshal error: %v, items: %v", err, item)
			errMsg := fmt.Sprintf("sender: sendSystemMetrics: failed to marshal value: %s for key: %s", item.ValueJson, item.Key)
			fs.logger.CaptureError(errMsg, e)
			continue
		}

		row["system."+item.Key] = val
	}

	// marshal the row
	line, err := json.Marshal(row)
	if err != nil {
		fs.logger.CaptureError("sender: sendSystemMetrics: failed to marshal system metrics", err)
		return
	}

	fs.addTransmit(processedChunk{
		fileType: EventsChunk,
		fileLine: string(line),
	})
}

func (fs *FileStream) streamPreempting(exitRecord *service.RunPreemptingRecord) {
	fs.addTransmit(processedChunk{
		Preempting: true,
	})
}

func (fs *FileStream) streamFilesUploaded(msg *service.FilesUploaded) {
	fs.addTransmit(processedChunk{
		Uploaded: msg.Files,
	})
}

func (fs *FileStream) streamFinish(exitRecord *service.RunExitRecord) {
	fs.addTransmit(processedChunk{
		Complete: &boolTrue,
		Exitcode: &exitRecord.ExitCode,
	})
}

func (fs *FileStream) streamStreamData(msg *service.StreamDataRecord) {
	row := make(map[string]interface{})
	for k, v := range msg.Items {
		switch x := v.StreamValueType.(type) {
		case *service.StreamValue_Int64Value:
			row[k] = x.Int64Value
		case *service.StreamValue_DoubleValue:
			row[k] = x.DoubleValue
		case *service.StreamValue_StringValue:
			row[k] = x.StringValue
		}
	}
	if fs.streamTableClientId != "" {
		row["_client_id"] = fs.streamTableClientId
	}
	// fmt.Printf("SEND %+v\n", row)
	jsonBytes, err := json.Marshal(row)
	if err != nil {
		panic("badness")
	}
	line := string(jsonBytes)
	// fmt.Printf("LINE %+v\n", line)

	if err != nil {
		fs.logger.CaptureFatalAndPanic("json unmarshal error", err)
	}
	fs.addTransmit(processedChunk{
		fileType: HistoryChunk,
		fileLine: line,
	})
}