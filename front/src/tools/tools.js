export function shortenUuid(uuid, startLen = 4, endLen = 6) {
    const str = String(uuid);
    if (startLen + endLen >= str.length)
        return str;
    const start = str.substring(0, startLen);
    const end = str.substring(str.length - endLen);
    return `${start}…${end}`;
}
//# sourceMappingURL=tools.js.map