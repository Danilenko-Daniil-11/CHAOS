// Notification functions for CHAOS
function showNotification(message, type = 'info') {
    // Basic notification using alert for now
    alert(message);
}

function showSuccess(message) {
    showNotification(message, 'success');
}

function showError(message) {
    showNotification(message, 'error');
}

function showWarning(message) {
    showNotification(message, 'warning');
}
