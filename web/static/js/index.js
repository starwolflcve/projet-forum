const postList = document.querySelector('#post-list');
const authorFilter = document.querySelector('#filter-author');
const searchFilter = document.querySelector('#filter-search');
const sortFilter = document.querySelector('#filter-sort');
const postForm = document.querySelector('#post-form');
const commentForm = document.querySelector('#comment-form');
const postDetail = document.querySelector('#post-detail');
const selectedPostLabel = document.querySelector('#selected-post-label');
const categoryInput = document.querySelector('#post-categories');
const platformsButtonsContainer = document.querySelector('#platforms-buttons');
const currentPlatformTitle = document.querySelector('#current-platform-title');
const createPostModal = document.querySelector('#create-post-modal');

let categories = [];
let currentPostId = null;
let selectedCategorySlug = '';

async function fetchCategories() {
  try {
    const response = await fetch('/api/categories');
    if (!response.ok) throw new Error('Impossible de charger les catégories');
    categories = await response.json();
    renderPlatformButtons();
    renderCategoryOptions();
  } catch (err) {
    console.error(err);
  }
}

function renderPlatformButtons() {
  let html = '<button class="platform-btn active" onclick="selectPlatform(\'\', \'Tous les posts\')">📍 Tous les posts</button>';
  
  categories.forEach(cat => {
    html += `<button class="platform-btn" onclick="selectPlatform('${cat.slug}', '${cat.name}')">${getEmojiForCategory(cat.name)} ${cat.name}</button>`;
  });
  
  platformsButtonsContainer.innerHTML = html;
}

function getEmojiForCategory(name) {
  const emojiMap = {
    'HackTheBox': '🎯',
    'Root Me': '🌳',
    'TryHackMe': '🎓',
    'PicoCTF': '🚩',
    'OverTheWire': '🔗',
    'OWASP WebGoat': '🍷',
    'Web Exploitation': '🕷️',
    'Reverse Engineering': '⚙️',
    'Cryptographie': '🔐',
    'Forensics': '🔬',
    'Steganographie': '🎨',
    'Recon & OSINT': '🔍',
    'Buffer Overflow': '💥',
    'Malware Analysis': '🦠',
  };
  return emojiMap[name] || '🏷️';
}

function renderCategoryOptions() {
  categoryInput.innerHTML = '<option value="">-- Choisir une plateforme --</option>' + 
    categories.map(cat => `<option value="${cat.id}">${getEmojiForCategory(cat.name)} ${cat.name}</option>`).join('');
}

function selectPlatform(slug, name) {
  selectedCategorySlug = slug;
  currentPlatformTitle.textContent = name;
  
  // Mettre à jour les boutons actifs
  document.querySelectorAll('.platform-btn').forEach(btn => {
    if (btn.getAttribute('onclick').includes(`'${slug}'`)) {
      btn.classList.add('active');
    } else {
      btn.classList.remove('active');
    }
  });
  
  fetchPosts();
}

async function fetchPosts() {
  const params = new URLSearchParams();
  if (selectedCategorySlug) params.set('category', selectedCategorySlug);
  if (authorFilter.value) params.set('author', authorFilter.value.trim());
  if (searchFilter.value) params.set('q', searchFilter.value);
  if (sortFilter.value) params.set('sort_by', sortFilter.value);

  const response = await fetch(`/api/posts?${params.toString()}`);
  const posts = await response.json();
  renderPostList(posts);
}

function renderPostList(posts) {
  if (!posts || posts.length === 0) {
    postList.innerHTML = '<p style="padding: 20px; text-align: center; color: #888;">Aucun post trouvé pour ces filtres.</p>';
    return;
  }

  postList.innerHTML = posts.map(post => `
    <article class="post-card">
      <div class="post-meta">
        <span class="tag">#${post.id}</span>
        <span>👤 ${escapeHtml(post.author_name || post.user_id)}</span>
        <span>📅 ${new Date(post.created_at).toLocaleString('fr-FR')}</span>
      </div>
      <h3>${escapeHtml(post.title)}</h3>
      <p>${escapeHtml(post.content).slice(0, 220)}${post.content.length > 220 ? '...' : ''}</p>
      <div class="row" style="margin-top:18px;">
        <button class="secondary small" onclick="selectPost(${post.id})">📖 Voir</button>
        <button class="secondary small" onclick="reactPost(${post.id}, 'like')">👍 Like</button>
        <button class="secondary small" onclick="reactPost(${post.id}, 'dislike')">👎 Dislike</button>
      </div>
    </article>
  `).join('');
}

async function loadPostDetail(postId) {
  currentPostId = postId;
  selectedPostLabel.textContent = `Post sélectionné #${postId}`;

  const response = await fetch(`/api/posts/detail?id=${postId}`);
  if (!response.ok) {
    postDetail.innerHTML = '<p>Impossible de charger les détails du post.</p>';
    return;
  }

  const data = await response.json();
  const { post, comments, likes, dislikes } = data;

  postDetail.innerHTML = `
    <div class="detail-panel">
      <div class="post-meta">
        <span class="tag">#${post.id}</span>
        <span>👤 ${escapeHtml(post.author_name || post.user_id)}</span>
        <span>📅 ${new Date(post.created_at).toLocaleString('fr-FR')}</span>
      </div>
      <h3>${escapeHtml(post.title)}</h3>
      <p>${escapeHtml(post.content)}</p>
      <div class="post-meta" style="margin-top:18px;">
        <span>👍 Likes: ${likes}</span>
        <span>👎 Dislikes: ${dislikes}</span>
      </div>
    </div>
    <section class="panel panel-dark">
      <h3>💬 Commentaires (${comments.length})</h3>
      ${comments.length ? comments.map(renderComment).join('') : '<p>Aucun commentaire pour l\'instant.</p>'}
    </section>
  `;
}

function renderComment(comment) {
  return `
    <div class="comment-card">
      <div class="comment-meta">
        <span>💬 #${comment.id}</span>
        <span>👤 ${comment.user_id}</span>
        <span>📅 ${new Date(comment.created_at).toLocaleString('fr-FR')}</span>
      </div>
      <p>${escapeHtml(comment.content)}</p>
      <div class="row" style="margin-top:16px;">
        <button class="secondary small" onclick="reactComment(${comment.id}, 'like')">👍 Like</button>
        <button class="secondary small" onclick="reactComment(${comment.id}, 'dislike')">👎 Dislike</button>
      </div>
    </div>
  `;
}

function selectPost(postId) {
  loadPostDetail(postId);
}

async function reactPost(postId, reactionType) {
  await fetch('/api/posts/react', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ target_id: postId, reaction_type: reactionType }),
  });
  if (currentPostId === postId) {
    loadPostDetail(postId);
  }
}

async function reactComment(commentId, reactionType) {
  await fetch('/api/comments/react', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ target_id: commentId, reaction_type: reactionType }),
  });
  if (currentPostId) {
    loadPostDetail(currentPostId);
  }
}

function toggleCreatePost() {
  createPostModal.classList.toggle('active');
  if (createPostModal.classList.contains('active')) {
    postForm.reset();
  }
}

postForm.addEventListener('submit', async event => {
  event.preventDefault();

  const categoryId = document.querySelector('#post-categories').value;
  if (!categoryId) {
    alert('Veuillez choisir une plateforme/catégorie');
    return;
  }

  const body = {
    title: document.querySelector('#post-title').value,
    content: document.querySelector('#post-content').value,
    image_path: document.querySelector('#post-image').value,
    visibility: document.querySelector('#post-visibility').value,
    moderation_status: document.querySelector('#post-moderation').value,
    author_name: document.querySelector('#post-author-name').value.trim(),
    category_ids: [Number(categoryId)],
  };

  const userId = Number(document.querySelector('#post-user-id').value) || 1;
  
  try {
    const response = await fetch(`/api/posts/create?user_id=${userId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    
    if (response.ok) {
      alert('Post créé avec succès! 🎉');
      toggleCreatePost();
      postForm.reset();
      fetchPosts();
    } else {
      alert('Erreur lors de la création du post');
    }
  } catch (err) {
    console.error(err);
    alert('Erreur: ' + err.message);
  }
});

commentForm.addEventListener('submit', async event => {
  event.preventDefault();
  if (!currentPostId) {
    alert('Sélectionne d\'abord un post.');
    return;
  }

  const body = {
    post_id: currentPostId,
    content: document.querySelector('#comment-content').value,
  };

  const userId = Number(document.querySelector('#comment-user-id').value) || 1;
  
  try {
    const response = await fetch(`/api/comments/create?user_id=${userId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    
    if (response.ok) {
      alert('Commentaire ajouté! 📝');
      commentForm.reset();
      loadPostDetail(currentPostId);
    } else {
      alert('Erreur lors de l\'ajout du commentaire');
    }
  } catch (err) {
    console.error(err);
    alert('Erreur: ' + err.message);
  }
});

[authorFilter, searchFilter, sortFilter].forEach(element => {
  element.addEventListener('change', fetchPosts);
});

function escapeHtml(text) {
  return String(text)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

window.addEventListener('DOMContentLoaded', async () => {
  await fetchCategories();
  fetchPosts();
});
