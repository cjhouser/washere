async function getSignatures() {
    const signatureList = document.getElementById('signatures');
    const lastSignature = signatureList.lastElementChild();
    const response = await fetch(`http://192.168.0.252:32223/signatures?page=${lastSignature.id}`);
    const signatures = response.json().signatures
    signatures.forEach(signature => {
        const signatureItem = document.createElement('li');
        signatureItem.id = signature.id
        signatureItem.textContent = signature.text
        signatureList.appendChild(signatureItem);
    });
}