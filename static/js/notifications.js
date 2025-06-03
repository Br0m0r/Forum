async function pollNotifCount() {
    console.log('pollNotifCount: Starting notification count check');
    try {
      console.log('pollNotifCount: Fetching notification count from server');
      const res = await fetch('/notifications/count');
      
      if (!res.ok) {
        console.warn('pollNotifCount: Server returned non-OK status:', res.status, res.statusText);
        return;
      }
      
      const data = await res.json();
      const { count } = data;
      console.log('pollNotifCount: Received notification count:', count);
      
      const btn = document.querySelector('.notif-button');
      if (!btn) {
        console.warn('pollNotifCount: Notification button not found in DOM');
        return;
      }
      console.log('pollNotifCount: Found notification button element');
      
      let badge = btn && btn.querySelector('.notif-badge');
      console.log('pollNotifCount: Current badge status:', badge ? 'exists' : 'does not exist');
  
      if (count > 0) {
        console.log('pollNotifCount: Showing badge with count', count);
        if (!badge) {
          console.log('pollNotifCount: Creating new badge element');
          badge = document.createElement('span');
          badge.className = 'notif-badge';
          btn.appendChild(badge);
          console.log('pollNotifCount: Badge element appended to button');
        }
        badge.textContent = count;
        console.log('pollNotifCount: Badge text updated to', count);
      } else if (badge) {
        console.log('pollNotifCount: Removing badge as count is zero');
        badge.remove();
        console.log('pollNotifCount: Badge removed from DOM');
      }
    } catch (err) {
      console.error('pollNotifCount: Error during notification polling:', err);
    }
    console.log('pollNotifCount: Notification check completed');
  }
  
  document.addEventListener('DOMContentLoaded', () => {
    console.log('Notifications: DOM loaded, initializing notification polling');
    pollNotifCount();
    console.log('Notifications: Setting up polling interval (15 seconds)');
    setInterval(pollNotifCount, 15000); // every 15 seconds
  });