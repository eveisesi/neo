export default ({ app }) => {
    const client = app.apolloProvider.defaultClient

    client.wsClient.lazy = true
    client.wsClient.reconnect = true
}