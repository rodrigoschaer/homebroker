## Technologies
- Docker
- GoLang
- Nest.js
- Next.js / React
- Apache Kafka

## Architecture

This application has 3 layers

- Frontend, Homebroker with Next.js
- Backend, Rest layer with Nest.js
- Stock System (pricing and dealing) layer with GoLang
(Server Event and messaging with Apache Kafka)

<img src="https://github.com/rodrigoschaer/homebroker/assets/70034234/550e0636-bfea-4c27-a4c1-f58b7aa79811">

Everything happens in real time with Server Sent Event (SSE) communication in FE and BE layers.

## Stock System (Go)
Implements buy and sell orders algorythms:

- High efficency and performatic, using `goroutines` to process data in multi-thread operations
- Works around with IN MEMORY operations
- Main data processed in a `Book` entity, allocating in maps and heaps

How it works:
- Every time a buy order matches a sell order, a transaction event is generated
- This transaction is published through Kafka in a Json format
- There is no use of a database, it all happens in real time and in memory, due to time and design complexity (this will be added later, with a Transaction history feature)

<img src="https://github.com/rodrigoschaer/homebroker/assets/70034234/19819ec8-8ea2-4188-a9b3-b2979ce063de" width="50%" height="50%">

Stack usage:
- Uses channels to deal with thread communication to prevent race condition (multiple threads changing the same data)
- Every order will be processed in a single channel, so they are registered in the same Book.

Architecture:
- Orders (BUY/SELL) are received (funneled) from Apache Kafka, in "Orders" topic, and are saved in Order channel
- An asynchronous channel, called trade channel (further Book.Trade() function) will implement the matchings and will fire the transactions to next channel
- The transactions channel will fire events to Kafka, inside "Transactions" topic
- Everything is wrapped in a docker container

<img src="https://github.com/rodrigoschaer/homebroker/assets/70034234/17e6c78c-6a01-4c79-952b-96992f7414c9" width="75%" height="75%">

## Backend Layer with Nest.js

<img src="https://github.com/rodrigoschaer/homebroker/assets/70034234/92341e15-8029-43cf-9927-bb206d92495a" width="50%" height="50%">


