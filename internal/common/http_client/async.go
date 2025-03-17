package http_client

type ClientGenerator func() (Client, error)

func PrepareClientGenerator(
	reqFunc ClientRequestFunc,
	resFunc ClientResponseFunc,
) ClientGenerator {
	return func() (Client, error) {
		return SendDefault(reqFunc, resFunc)
	}
}

type ClientAsync struct {
	Client Client
	Err    error
}

func AsyncDefault(
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
