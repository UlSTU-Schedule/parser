// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package types

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes(in *jlexer.Lexer, out *Schedule) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "weeks":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('[')
				v1 := 0
				for !in.IsDelim(']') {
					if v1 < 2 {
						easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes1(in, &(out.Weeks)[v1])
						v1++
					} else {
						in.SkipRecursive()
					}
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes(out *jwriter.Writer, in Schedule) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"weeks\":"
		out.RawString(prefix[1:])
		out.RawByte('[')
		for v2 := range in.Weeks {
			if v2 > 0 {
				out.RawByte(',')
			}
			easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes1(out, (in.Weeks)[v2])
		}
		out.RawByte(']')
	}
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Schedule) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Schedule) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes(l, v)
}
func easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes1(in *jlexer.Lexer, out *Week) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "number":
			out.Number = int(in.Int())
		case "days":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('[')
				v3 := 0
				for !in.IsDelim(']') {
					if v3 < 7 {
						easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes2(in, &(out.Days)[v3])
						v3++
					} else {
						in.SkipRecursive()
					}
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes1(out *jwriter.Writer, in Week) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"number\":"
		out.RawString(prefix[1:])
		out.Int(int(in.Number))
	}
	{
		const prefix string = ",\"days\":"
		out.RawString(prefix)
		out.RawByte('[')
		for v4 := range in.Days {
			if v4 > 0 {
				out.RawByte(',')
			}
			easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes2(out, (in.Days)[v4])
		}
		out.RawByte(']')
	}
	out.RawByte('}')
}
func easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes2(in *jlexer.Lexer, out *Day) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "week_number":
			out.WeekNumber = int(in.Int())
		case "lessons":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('[')
				v5 := 0
				for !in.IsDelim(']') {
					if v5 < 8 {
						easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes3(in, &(out.Lessons)[v5])
						v5++
					} else {
						in.SkipRecursive()
					}
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes2(out *jwriter.Writer, in Day) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"week_number\":"
		out.RawString(prefix[1:])
		out.Int(int(in.WeekNumber))
	}
	{
		const prefix string = ",\"lessons\":"
		out.RawString(prefix)
		out.RawByte('[')
		for v6 := range in.Lessons {
			if v6 > 0 {
				out.RawByte(',')
			}
			easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes3(out, (in.Lessons)[v6])
		}
		out.RawByte(']')
	}
	out.RawByte('}')
}
func easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes3(in *jlexer.Lexer, out *Lesson) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "sub_lessons":
			if in.IsNull() {
				in.Skip()
				out.SubLessons = nil
			} else {
				in.Delim('[')
				if out.SubLessons == nil {
					if !in.IsDelim(']') {
						out.SubLessons = make([]SubLesson, 0, 0)
					} else {
						out.SubLessons = []SubLesson{}
					}
				} else {
					out.SubLessons = (out.SubLessons)[:0]
				}
				for !in.IsDelim(']') {
					var v7 SubLesson
					easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes4(in, &v7)
					out.SubLessons = append(out.SubLessons, v7)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes3(out *jwriter.Writer, in Lesson) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"sub_lessons\":"
		out.RawString(prefix[1:])
		if in.SubLessons == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v8, v9 := range in.SubLessons {
				if v8 > 0 {
					out.RawByte(',')
				}
				easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes4(out, v9)
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}
func easyjson6601e8cdDecodeGithubComUlstuScheduleParserTypes4(in *jlexer.Lexer, out *SubLesson) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "duration":
			out.Duration = Duration(in.Int())
		case "type":
			out.Type = LessonType(in.Int())
		case "group":
			out.Group = string(in.String())
		case "name":
			out.Name = string(in.String())
		case "teacher":
			out.Teacher = string(in.String())
		case "room":
			out.Room = string(in.String())
		case "practice":
			out.Practice = string(in.String())
		case "sub_group":
			out.SubGroup = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6601e8cdEncodeGithubComUlstuScheduleParserTypes4(out *jwriter.Writer, in SubLesson) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"duration\":"
		out.RawString(prefix[1:])
		out.Int(int(in.Duration))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.Int(int(in.Type))
	}
	{
		const prefix string = ",\"group\":"
		out.RawString(prefix)
		out.String(string(in.Group))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"teacher\":"
		out.RawString(prefix)
		out.String(string(in.Teacher))
	}
	{
		const prefix string = ",\"room\":"
		out.RawString(prefix)
		out.String(string(in.Room))
	}
	{
		const prefix string = ",\"practice\":"
		out.RawString(prefix)
		out.String(string(in.Practice))
	}
	{
		const prefix string = ",\"sub_group\":"
		out.RawString(prefix)
		out.String(string(in.SubGroup))
	}
	out.RawByte('}')
}
