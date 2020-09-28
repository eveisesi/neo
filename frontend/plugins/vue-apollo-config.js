export default () => {

    const output = {
        httpEndpoint: process.env.apiUrl,
        wsEndpoint: process.env.wssUrl,
    }

    return output
}