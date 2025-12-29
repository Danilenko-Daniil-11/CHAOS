// Common JavaScript functions for CHAOS
function showAlert(message, type = 'info') {
    // Basic alert function
    alert(message);
}

function confirmAction(message) {
    return confirm(message);
}

// Utility functions
function encodeBase64(str) {
    return btoa(str);
}

function decodeBase64(str) {
    return atob(str);
}
