package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func defaultHandleError(_ context.Context, err error) []zap.Field {
	if err == nil {
		return nil
	}
	fields := []zap.Field{zap.Error(err)}
	if st, ok := status.FromError(err); ok {
		fields = append(fields, zap.String(fieldGRPCErrorMessage, st.Message()))
		details := st.Details()
		if len(details) > 0 {
			fields = append(fields, zap.Array(fieldGRPCErrorDetails, &errorDetailsObjectMarshaler{details}))
		}
	}
	return fields
}

type errorDetailsObjectMarshaler struct {
	details []any
}

func (e errorDetailsObjectMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, detail := range e.details {
		switch detail.(type) {
		case *errdetails.BadRequest, *errdetails.QuotaFailure, *errdetails.RequestInfo, *errdetails.ResourceInfo, *errdetails.DebugInfo, *errdetails.Help, *errdetails.LocalizedMessage, *errdetails.PreconditionFailure, *errdetails.RetryInfo, *errdetails.ErrorInfo:
			encoder.AppendObject(&ErrDetailObjectMarshaler{detail.(proto.Message)})
		default:
			encoder.AppendReflected(detail)
		}
	}
	return nil
}

type ErrDetailObjectMarshaler struct {
	detail proto.Message
}

func (e ErrDetailObjectMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("$type", string(proto.MessageName(e.detail).Name()))
	switch v := e.detail.(type) {
	case *errdetails.BadRequest:
		encoder.AddArray("field_violations", &fieldViolationsObjectMarshaler{v.FieldViolations})
	case *errdetails.QuotaFailure:
		encoder.AddArray("violations", &quotaViolationsObjectMarshaler{v.Violations})
	case *errdetails.RequestInfo:
		encoder.AddString("request_id", v.RequestId)
		encoder.AddString("serving_data", v.ServingData)
	case *errdetails.ResourceInfo:
		encoder.AddString("resource_type", v.ResourceType)
		encoder.AddString("resource_name", v.ResourceName)
		encoder.AddString("owner", v.Owner)
		encoder.AddString("description", v.Description)
	case *errdetails.DebugInfo:
		encoder.AddArray("stack_entries", &stringsArrayMarshaler{v.StackEntries})
		encoder.AddString("detail", v.Detail)
	case *errdetails.Help:
		encoder.AddArray("links", &linksObjectMarshaler{v.Links})
	case *errdetails.LocalizedMessage:
		encoder.AddString("locale", v.Locale)
		encoder.AddString("message", v.Message)
	case *errdetails.PreconditionFailure:
		encoder.AddArray("violations", &preconditionViolationsObjectMarshaler{v.Violations})
	case *errdetails.RetryInfo:
		encoder.AddString("retry_delay", v.RetryDelay.String())
	case *errdetails.ErrorInfo:
		encoder.AddString("reason", v.Reason)
		encoder.AddString("domain", v.Domain)
		for k, v := range v.Metadata {
			encoder.AddString(fmt.Sprintf("metadata.%s", k), v)
		}
	}
	return nil
}

type quotaViolationsObjectMarshaler struct {
	violations []*errdetails.QuotaFailure_Violation
}

func (q quotaViolationsObjectMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, violation := range q.violations {
		encoder.AppendObject(&quotaViolationObjectMarshaler{violation})
	}
	return nil
}

type quotaViolationObjectMarshaler struct {
	violation *errdetails.QuotaFailure_Violation
}

func (q quotaViolationObjectMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("subject", q.violation.Subject)
	encoder.AddString("description", q.violation.Description)
	return nil
}

type fieldViolationsObjectMarshaler struct {
	fieldViolations []*errdetails.BadRequest_FieldViolation
}

func (f fieldViolationsObjectMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, fieldViolation := range f.fieldViolations {
		encoder.AppendObject(&fieldViolationObjectMarshaler{fieldViolation})
	}
	return nil
}

type fieldViolationObjectMarshaler struct {
	fieldViolation *errdetails.BadRequest_FieldViolation
}

func (f fieldViolationObjectMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("field", f.fieldViolation.Field)
	encoder.AddString("description", f.fieldViolation.Description)
	return nil
}

type stringsArrayMarshaler struct {
	arr []string
}

func (s stringsArrayMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, str := range s.arr {
		encoder.AppendString(str)
	}
	return nil
}

type linksObjectMarshaler struct {
	links []*errdetails.Help_Link
}

func (l linksObjectMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, link := range l.links {
		encoder.AppendObject(&linkObjectMarshaler{link})
	}
	return nil
}

type linkObjectMarshaler struct {
	link *errdetails.Help_Link
}

func (l linkObjectMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("description", l.link.Description)
	encoder.AddString("url", l.link.Url)
	return nil
}

type preconditionViolationsObjectMarshaler struct {
	violations []*errdetails.PreconditionFailure_Violation
}

func (p preconditionViolationsObjectMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, violation := range p.violations {
		encoder.AppendObject(&preconditionViolationObjectMarshaler{violation})
	}
	return nil
}

type preconditionViolationObjectMarshaler struct {
	violation *errdetails.PreconditionFailure_Violation
}

func (p preconditionViolationObjectMarshaler) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("type", p.violation.Type)
	encoder.AddString("subject", p.violation.Subject)
	encoder.AddString("description", p.violation.Description)
	return nil
}
