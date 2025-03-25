package async

import "github.com/sakamotoryou/api-agg-two/internal/common/http_client/Service/client"

type ClientGenerator func() (client.Client, error)

func PrepareClientGenerator(
	reqFunc client.ClientRequestFunc,
	resFunc client.ClientResponseFunc,
) ClientGenerator {
	return func() (client.Client, error) {
		return client.SendDefault(reqFunc, resFunc)
	}
}

type ClientAsync struct {
	Client client.Client
	Err    error
}

func Async(
	done <-chan struct{},
	clientFunc ...ClientGenerator,
) <-chan ClientAsync {
	clientStream := make(chan ClientAsync)
	go func() {
		defer close(clientStream)
		for _, send := range clientFunc {
			client, err := send()
			c := ClientAsync{
				Client: client,
				Err:    err,
			}

			select {
			case <-done:
				return
			case clientStream <- c:
			}
		}
	}()

	return clientStream
}

func Do(clientFunc ...ClientGenerator) <-chan ClientAsync {
  done := make(chan struct{})
  defer close(done)

  async_client := Async(done, clientFunc...)
  for client := range async_client {

  }
}
