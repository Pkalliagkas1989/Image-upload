const forumContainer = document.getElementById('forumContainer');
const postTemplate = document.getElementById('post-template');

// Endpoint used to verify the session and fetch the CSRF token
const sessionVerifyURL = 'http://localhost:8080/forum/api/session/verify';

// In-memory storage for the CSRF token
let csrfTokenFromResponse = null;

// Helper to load a fresh CSRF token from the backend
async function loadCSRFTokenFromSession() {
  try {
    const resp = await fetch(sessionVerifyURL, { credentials: 'include' });
    if (!resp.ok) throw new Error('Session not valid');
    const data = await resp.json();
    return data.csrf_token || data.CSRFToken;
  } catch (err) {
    console.warn('Failed to load CSRF token:', err);
    return null;
  }
}

window.addEventListener('DOMContentLoaded', async () => {
  // Load CSRF token when the page is loaded
  csrfTokenFromResponse = await loadCSRFTokenFromSession();
  fetchCreatedPosts();
});

async function fetchCreatedPosts() {
  try {
    const resp = await fetch('http://localhost:8080/forum/api/user/posts', {
      credentials: 'include',
    });

    if (!resp.ok) {
      const err = await resp.json();
      throw new Error(err.message || 'Failed to load created posts');
    }

    const posts = await resp.json();
    renderCreatedPosts(posts);
  } catch (err) {
    console.error(`Error: ${err.message}`);
    forumContainer.textContent = 'You have not created any posts yet.';
  }
}

function renderCreatedPosts(posts) {
  forumContainer.innerHTML = '';

  if (!posts.length) {
    forumContainer.textContent = 'You have not created any posts yet.';
    return;
  }

  posts.forEach(post => {
    const node = postTemplate.content.cloneNode(true);
    const postEl = node.querySelector('.post');
    if (post.thumbnail_url) {
      const img = document.createElement('img');
      img.src = post.thumbnail_url;
      img.alt = 'Post thumbnail';
      img.className = 'post-thumb';
      postEl.insertBefore(img, postEl.firstChild);
    }

    node.querySelector('.post-header').textContent = post.username || 'You';
    node.querySelector('.post-title').textContent = post.title;
    node.querySelector('.post-content').textContent = post.content;
    node.querySelector('.post-time').textContent = new Date(post.created_at).toLocaleString();

    const likeCount = (post.reactions || []).filter(r => r.reaction_type === 1).length;
    const dislikeCount = (post.reactions || []).filter(r => r.reaction_type === 2).length;

    node.querySelector('.like-count').textContent = likeCount;
    node.querySelector('.dislike-count').textContent = dislikeCount;

    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);
    const commentContainer = document.createElement('span');
    commentContainer.className = 'comment-count';
    commentContainer.innerHTML = `💬 ${commentCount}`;
    node
      .querySelector('.like-count')
      .parentNode.appendChild(commentContainer);

    const wrapper = document.createElement('a');
    wrapper.href = `/user/post?id=${post.id}`;
    wrapper.className = 'post-link';
    wrapper.appendChild(node);

    const editBtn = wrapper.querySelector('.edit-post');
    const deleteBtn = wrapper.querySelector('.delete-post');

    editBtn.addEventListener('click', async e => {
      e.preventDefault();

      // Ensure we have a CSRF token before making the request
      if (!csrfTokenFromResponse) {
        csrfTokenFromResponse = await loadCSRFTokenFromSession();
        if (!csrfTokenFromResponse) {
          alert('Session expired. Please log in again.');
          return;
        }
      }

      const title = prompt('Edit title', post.title);
      if (title === null) return;
      const content = prompt('Edit content', post.content);
      if (content === null) return;
      const body = {
        post_id: post.id,
        title,
        content,
        category_ids: post.categories ? post.categories.map(c => c.id) : [],
      };
      try {
        const resp = await fetch('http://localhost:8080/forum/api/posts/update', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfTokenFromResponse,
          },
          body: JSON.stringify(body),
        });
        if (!resp.ok) throw new Error('failed');
        fetchCreatedPosts();
      } catch (err) {
        alert('Failed to update post');
      }
    });

    deleteBtn.addEventListener('click', async e => {
      e.preventDefault();
      if (!confirm('Delete this post?')) return;

      // Ensure CSRF token
      if (!csrfTokenFromResponse) {
        csrfTokenFromResponse = await loadCSRFTokenFromSession();
        if (!csrfTokenFromResponse) {
          alert('Session expired. Please log in again.');
          return;
        }
      }

      try {
        const resp = await fetch('http://localhost:8080/forum/api/posts/delete', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfTokenFromResponse,
          },
          body: JSON.stringify({ post_id: post.id }),
        });
        if (!resp.ok) throw new Error('failed');
        fetchCreatedPosts();
      } catch (err) {
        alert('Failed to delete post');
      }
    });

    forumContainer.appendChild(wrapper);
  });
}
