# investor-chat
investor-chat is a real-time chat application that runs in the browser.

## Preview
https://github.com/ap-pauloafonso/investor-chat/assets/26802978/f3de8351-894e-42c3-9a7f-e1e43dfd3d19

## Features
* Login/Register user
* Real time chat
* Multiple Channels
* Messages are archived in the database 
* Chat-bot to get the current stock price: `/stock [STOCK_CODE]` - example `/stock=aapl.us`

## Project 
![Diagram](media/diagram.png)

### Backend
The Backend was built using Go, and is architected for high performance and scalability, featuring several key components:

* WebSocket Communication: Using WebSockets to maintain open connections with clients, ensuring a faster chat experience compared to standard REST APIs.
* Scalability: The architecture supports horizontal scalability with a reverse-proxy/load balancer (HAProxy) to distribute requests evenly across WebSocket servers. This can be expanded by spinning up additional WebSocketServer instances if needed.
    * This is achieved through the implementation of a 'sticky session' concept, where the balancer assigns a unique session cookie to the client. This cookie ensures that the user remains connected to a specific server until the session is terminated
* BotServer: Listens to messages requesting stock information, processes them, and places them back in the queue for WebSocketServers to consume and broadcast to all connections.
* Pub/Sub Mechanism: Employing RabbitMQ for efficient communication between WebSocketServers and the BotServer.
* Data Persistence: Since we want to have a history of data persisted, a consumer is set up to listen to new messages and write them to PostgreSQL. While non-relational databases like Cassandra could be used for faster read/write operations, PostgreSQL was chosen for simplicity.

### Frontend (clients)
* Built with React and tailwindcss.

## Running it with Docker and Testing
* Make sure that you don't have any other docker containers running
* Run `docker-compose up --build`
* Go to http://localhost
* For testing with a second account, open an anonymous tab or another browser, as sessions are managed through cookies.



## Completed Bonus
- [x] Have more than one chatroom.

## Todo
- write comments
- write more tests
