export default (ctx) => {

    const output = {
        httpEndpoint: ctx.env.apiURL,
        wsEndpoint: ctx.env.wssURL,
    }

    return output
}