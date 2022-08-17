package command

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func ExampleCommand() {

}

func TestCommand(t *testing.T) {
	convey.Convey("TestCommand", t, func() {
		convey.Convey("exec date", func() {
			cmd := &Command{}
			err := cmd.WithWorkSpace("D:/WorkSpace/gitee/example").
				WithCmdStr("D:/WorkSpace/gitee/example/example.exe").
				WithTimeout(time.Second * 3).
				Run(context.Background())
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("exec example timeout", func() {
			cmd := &Command{}
			err := cmd.WithWorkSpace("D:/WorkSpace/gitee/example").
				WithCmdStr("D:/WorkSpace/gitee/example/example.exe").
				WithTimeout(time.Second * 6).
				Run(context.Background())
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
