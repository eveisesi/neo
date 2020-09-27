export default function head(name, type) {

    let out = name
    if (type != undefined) {
        out = out + " || " + type
    }

    return out + " || New Edem Obituary"
}