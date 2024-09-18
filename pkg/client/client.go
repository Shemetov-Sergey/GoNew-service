package client

import (
	"fmt"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/pb/comment"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/config"
	"google.golang.org/grpc"
)

type ServiceClient struct {
	Client comment.CommentServiceClient
}

func InitServiceClient(c *config.Config) comment.CommentServiceClient {
	// using WithInsecure() because no SSL running
	cc, err := grpc.Dial(c.CommentSvcUrl, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect:", err)
	}

	return comment.NewCommentServiceClient(cc)
}
