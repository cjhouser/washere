async function getSignatures() {
    const signatureList = document.getElementById('signatures');
    const page = Math.floor(signatureList.childElementCount / 10);
    const response = await fetch('http://192.168.0.252:32223/signatures?' + new URLSearchParams({page: page}));
    const json = await response.json();
    const signatures = json.signatures;
    var documentFragment = document.createDocumentFragment();
    signatures.forEach(signature => {
        const signatureItem = document.createElement('li');
        signatureItem.textContent = signature.text;
        documentFragment.append(signatureItem);
    });
    signatureList.append(documentFragment);
    if (signatureList.childElementCount % 10) {
        button = document.getElementById('get-signatures-button');
        button.disabled = true;
    }
}