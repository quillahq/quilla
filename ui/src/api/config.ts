const getConfig = async () => {
    const response = await fetch('/v1/config');
    return response.json();
}

export {
    getConfig
}