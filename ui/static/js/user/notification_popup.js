const openBtn = document.getElementById('open-notifs');
const modal = document.getElementById('notif-modal');
const closeBtn = modal ? modal.querySelector('.close-btn') : null;
const list = document.getElementById('notif-list');
const template = document.getElementById('notif-item-template');
const markAllBtn = document.getElementById('mark-read-btn');
const deleteAllBtn = document.getElementById('delete-all-btn');
const notifLink = document.getElementById('notifications-link');

// Endpoint used to verify the session and fetch the CSRF token
const sessionVerifyURL = 'http://localhost:8080/forum/api/session/verify';
// In-memory storage for the CSRF token
let csrfTokenFromResponse = null;

// Helper to load a CSRF token from the backend or cookie
async function loadCSRFTokenFromSession() {
  try {
    const csrfCookie = document.cookie
      .split('; ')
      .find(row => row.startsWith('csrf_token_frontend='));
    if (csrfCookie) {
      return csrfCookie.split('=')[1];
    }

    const resp = await fetch(sessionVerifyURL, { credentials: 'include' });
    if (!resp.ok) throw new Error('Session not valid');
    const data = await resp.json();
    return data.csrf_token || data.CSRFToken;
  } catch (err) {
    console.warn('Failed to load CSRF token:', err);
    return null;
  }
}

// check for unread notifications and highlight button
async function checkUnread() {
  try {
    const resp = await fetch('http://localhost:8080/forum/api/user/notifications', {
      credentials: 'include'
    });
    if (!resp.ok) throw new Error('failed');
    const data = await resp.json();
    const hasUnread = data.some(n => !n.is_read);
    toggleAlert(hasUnread);
  } catch (err) {
    console.error('Failed to check notifications', err);
  }
}

function toggleAlert(on) {
  if (openBtn) openBtn.classList.toggle('notification-alert', on);
  if (notifLink) notifLink.classList.toggle('notification-alert', on);
}

async function fetchNotifications() {
  try {
    const resp = await fetch('http://localhost:8080/forum/api/user/notifications', {
      credentials: 'include'
    });
    if (!resp.ok) throw new Error('failed');
    const data = await resp.json();
    renderList(data);
    const hasUnread = data.some(n => !n.is_read);
    toggleAlert(hasUnread);
  } catch (err) {
    console.error(err);
    list.textContent = 'Failed to load notifications';
  }
}

function renderList(notifs) {
  list.innerHTML = '';
  if (!notifs || notifs.length === 0) {
    list.textContent = 'No notifications';
    return;
  }
  notifs.forEach(n => {
    const node = template.content.cloneNode(true);
    const item = node.querySelector('.notification-item');
    if (!n.is_read) item.classList.add('unread');
    node.querySelector('.notif-text').textContent = formatMessage(n);
    node.querySelector('.notif-time').textContent = new Date(n.created_at).toLocaleString();

    const delBtn = node.querySelector('.delete-btn');
    delBtn.addEventListener('click', async () => {
      if (!csrfTokenFromResponse) {
        csrfTokenFromResponse = await loadCSRFTokenFromSession();
      }
      await fetch(`http://localhost:8080/forum/api/user/notifications/delete?id=${n.id}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: { 'X-CSRF-Token': csrfTokenFromResponse }
      });
      await fetchNotifications();
      checkUnread();
    });

    const markBtn = node.querySelector('.mark-btn');
    if (markBtn) {
      if (n.is_read) markBtn.classList.add('hidden');
      markBtn.addEventListener('click', async () => {
        if (!csrfTokenFromResponse) {
          csrfTokenFromResponse = await loadCSRFTokenFromSession();
        }
        await fetch('http://localhost:8080/forum/api/user/notifications/read', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfTokenFromResponse
          },
          body: JSON.stringify({ id: n.id })
        });
        await fetchNotifications();
        checkUnread();
      });
    }

    list.appendChild(node);
  });
}

function formatMessage(n) {
  const actor = n.actor_username || 'Someone';
  switch (n.action) {
    case 'like_post':
      return `${actor} liked your post`;
    case 'dislike_post':
      return `${actor} disliked your post`;
    case 'unlike_post':
      return `${actor} removed their like from your post`;
    case 'undislike_post':
      return `${actor} removed their dislike from your post`;
    case 'like_comment':
      return `${actor} liked your comment`;
    case 'dislike_comment':
      return `${actor} disliked your comment`;
    case 'unlike_comment':
      return `${actor} removed their like from your comment`;
    case 'undislike_comment':
      return `${actor} removed their dislike from your comment`;
    case 'comment':
      return `${actor} commented on your post`;
    default:
      return `${actor} did something`;
  }
}

if (openBtn && modal) {
  openBtn.addEventListener('click', async (e) => {
    e.preventDefault();
    await fetchNotifications();
    modal.classList.remove('hidden');
  });
}
if (closeBtn) {
  closeBtn.addEventListener('click', () => modal.classList.add('hidden'));
}
window.addEventListener('click', (e) => {
  if (e.target === modal) modal.classList.add('hidden');
});

if (markAllBtn) {
  markAllBtn.addEventListener('click', async () => {
    if (!csrfTokenFromResponse) {
      csrfTokenFromResponse = await loadCSRFTokenFromSession();
    }
    await fetch('http://localhost:8080/forum/api/user/notifications/read', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfTokenFromResponse },
      body: '{}'
    });
    await fetchNotifications();
    checkUnread();
  });
}
if (deleteAllBtn) {
  deleteAllBtn.addEventListener('click', async () => {
    if (!csrfTokenFromResponse) {
      csrfTokenFromResponse = await loadCSRFTokenFromSession();
    }
    await fetch('http://localhost:8080/forum/api/user/notifications/delete', {
      method: 'DELETE',
      credentials: 'include',
      headers: { 'X-CSRF-Token': csrfTokenFromResponse }
    });
    await fetchNotifications();
    checkUnread();
  });
}
window.addEventListener('DOMContentLoaded', async () => {
  csrfTokenFromResponse = await loadCSRFTokenFromSession();
  checkUnread();
});
