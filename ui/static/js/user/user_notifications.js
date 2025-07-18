const container = document.getElementById('notifContainer');
const template = document.getElementById('notif-template');

window.addEventListener('DOMContentLoaded', fetchNotifications);

async function fetchNotifications() {
  try {
    const resp = await fetch('http://localhost:8080/forum/api/user/notifications', {
      credentials: 'include',
    });
    if (!resp.ok) throw new Error('Failed to load notifications');
    const data = await resp.json();
    render(data);
  } catch (err) {
    console.error(err);
    container.textContent = 'Failed to load notifications.';
  }
}

function render(notifs) {
  container.innerHTML = '';
  if (!notifs || notifs.length === 0) {
    container.textContent = 'No notifications.';
    return;
  }
  notifs.forEach(n => {
    const node = template.content.cloneNode(true);
    node.querySelector('.notif-text').textContent = formatMessage(n);
    node.querySelector('.notif-time').textContent = new Date(n.created_at).toLocaleString();
    container.appendChild(node);
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
