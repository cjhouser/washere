async function getSignatures() {
    const signatureList = document.getElementById('signatures');
    const page = Math.floor(signatureList.childElementCount / 10);
    const response = await fetch('http://washere.com:32080/signatures?' + new URLSearchParams({page: page}));
    const signatures = await response.json();
    var documentFragment = document.createDocumentFragment();
    signatures.forEach(signature => {
        const signatureItem = document.createElement('li');
        signatureItem.textContent = signature.Text;
        documentFragment.append(signatureItem);
    });
    signatureList.append(documentFragment);
    if (signatureList.childElementCount % 10) {
        button = document.getElementById('get-signatures-button');
        button.disabled = true;
    }
}