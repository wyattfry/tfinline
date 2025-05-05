package inline

import (
	"io"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Lines map[string]Line

type Line struct {
	bar     *mpb.Bar
	message *string

	currentStatus Status
}

func NewLine(prog *mpb.Progress, address, msg string) *Line {
	l := Line{
		message:       &msg,
		currentStatus: StatusNone,
	}

	b := prog.New(
		1,
		mpb.SpinnerStyle(),
		mpb.PrependDecorators(
			decor.Name(address, decor.WCSyncSpaceR),
		),
		// Replaces `mpb.BarFillerOnComplete("msg")` to handle failures
		mpb.BarFillerMiddleware(func(filler mpb.BarFiller) mpb.BarFiller {
			return mpb.BarFillerFunc(func(w io.Writer, st decor.Statistics) error {
				if st.Completed {
					switch l.Status() {
					case StatusFailed:
						_, err := io.WriteString(w, "✗")
						return err
					default:
						_, err := io.WriteString(w, "✓")
						return err
					}
				}
				return filler.Fill(w, st)
			})
		}),
		mpb.AppendDecorators(
			decor.Any(func(_ decor.Statistics) string { return *l.message }),
		),
		mpb.BarWidth(1),
	)
	l.bar = b

	return &l
}

func (l *Line) MarkAsInProgress(msg string) {
	l.message = &msg
	l.currentStatus = StatusInProgress
}

func (l *Line) MarkAsFailed(msg string) {
	l.message = &msg
	l.currentStatus = StatusFailed
	l.bar.SetCurrent(1)
}

func (l *Line) MarkAsDone(msg string) {
	if msg != "" {
		l.message = &msg
	}
	l.currentStatus = StatusDone
	l.bar.SetCurrent(1)
}

func (l *Line) SetMessage(s string) {
	l.message = &s
}

func (l *Line) SetStatus(s Status) {
	l.currentStatus = s
}

func (l *Line) Status() Status {
	return l.currentStatus
}
