## Technologies
- Docker
- GoLang
- Nest.js
- Next.js / React
- Apache Kafka

## Architecture

This application has 4 layers

- Frontend, Homebroker with Next.js
- Backend, Rest layer with Nest.js
- Stock System (pricing and dealing) layer with GoLang
- Server Event and messaging with Apache Kafka

[ArchtectureImage]

Everything happens in real time with Server Sent Event (SSE) communication in FE and BE layers.

## Stock System (Go)
Implements buy and sell orders algorythms:

- High efficency and performatic
- Works around with IN MEMORY operations
- Allocates data mainly in heaps

[Buy/Sell Image]

How it works:
- Every time a buy order matches a sell order, a transaction event is generated
- This transaction is published through Kafka in a Json format
- There is no use of a database, it all happens in real time and in memory, due to time and design complexity (this will be added later, with a Transaction history feature)

Stack usage:
- Uses channels to deal with thread communication to prevent race condition (multiple threads changing the same data)
- Every order will be processed in a single channel, so they are registered in the same Book.

Architecture:
- Orders (BUY/SELL) are received (funneled) from Apache Kafka, in "Orders" topic, and are saved in Order channel
- An asynchronous channel, called trade channel (further Book.trade) will implement the matchings and will fire the transactions to next channel
- The transactions channel will fire events to Kafka, inside "Transactions" topic
- Everything is wrapped in a docker container
