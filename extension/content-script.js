// This script runs in the context of the webpage, so it can read its DOM.
// It will only collect data when requested by the popup
function collectPageData() {
    const url = window.location.href;
    const content = document.documentElement.innerText || "";

    // Send message to background script
    chrome.runtime.sendMessage({
        type: 'PAGE_DATA',
        data: { url, content }
    }, response => {
        if (chrome.runtime.lastError) {
            console.error('Error:', chrome.runtime.lastError);
            return;
        }
        if (response && response.success) {
            console.log('Data sent successfully');
        } else {
            console.error('Failed to send data');
        }
    });
}
