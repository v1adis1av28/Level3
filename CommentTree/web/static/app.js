class CommentApp {
    constructor() {
        this.baseUrl = window.location.origin;
        this.loadComments();
        this.setupEventListeners();
    }

    setupEventListeners() {
        document.getElementById('commentForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.submitComment();
        });

        document.getElementById('cancelReply').addEventListener('click', () => {
            this.cancelReply();
        });

        document.getElementById('search').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.searchComments();
            }
        });
    }

    async loadComments(parentId = 0) {
        try {
            const url = parentId === 0 ? 
                `${this.baseUrl}/comments` : 
                `${this.baseUrl}/comments?parent=${parentId}`;
            
            const response = await fetch(url);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            this.renderComments(data.comments);
        } catch (error) {
            console.error('Error loading comments:', error);
            alert('Ошибка загрузки комментариев: ' + error.message);
        }
    }

    renderComments(comments, container = document.getElementById('comments'), level = 0) {
        container.innerHTML = '';
        
        if (!comments || comments.length === 0) {
            container.innerHTML = '<p>Нет комментариев</p>';
            return;
        }
        
        comments.forEach(comment => {
            const commentDiv = document.createElement('div');
            commentDiv.className = 'comment';
            commentDiv.style.marginLeft = `${level * 20}px`;
            
            commentDiv.innerHTML = `
                <strong>${this.escapeHtml(comment.username)}</strong>
                <span>${new Date(comment.created_at).toLocaleString()}</span>
                <p>${this.escapeHtml(comment.text)}</p>
                <div class="actions">
                    <button onclick="app.replyTo(${comment.id}, '${this.escapeHtml(comment.username)}')">Ответить</button>
                    <button onclick="app.deleteComment(${comment.id})">Удалить</button>
                </div>
            `;

            container.appendChild(commentDiv);

            if (comment.children && comment.children.length > 0) {
                const childrenContainer = document.createElement('div');
                childrenContainer.className = 'children';
                container.appendChild(childrenContainer);
                this.renderComments(comment.children, childrenContainer, level + 1);
            }
        });
    }

    async submitComment() {
        const username = document.getElementById('username').value.trim();
        const text = document.getElementById('text').value.trim();
        const parentId = document.getElementById('parentId').value;

        if (!username || !text) {
            alert('Заполните все поля');
            return;
        }

        try {
            const response = await fetch(`${this.baseUrl}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    parent_id: parseInt(parentId),
                    username: username,
                    text: text
                })
            });

            if (response.ok) {
                document.getElementById('username').value = '';
                document.getElementById('text').value = '';
                if (parentId !== '0') {
                    this.cancelReply();
                }
                this.loadComments();
            } else {
                const errorData = await response.json();
                alert('Ошибка при добавлении комментария: ' + (errorData.error || response.statusText));
            }
        } catch (error) {
            console.error('Error submitting comment:', error);
            alert('Ошибка при добавлении комментария: ' + error.message);
        }
    }

    replyTo(parentId, username) {
        document.getElementById('parentId').value = parentId;
        document.getElementById('cancelReply').style.display = 'inline-block';
        document.getElementById('text').placeholder = `Ответ ${username}`;
        document.getElementById('text').focus();
    }

    cancelReply() {
        document.getElementById('parentId').value = '0';
        document.getElementById('cancelReply').style.display = 'none';
        document.getElementById('text').placeholder = 'Текст комментария';
    }

    async deleteComment(id) {
        if (!confirm('Удалить комментарий и все ответы?')) return;

        try {
            const response = await fetch(`${this.baseUrl}/comments/${id}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.loadComments();
            } else {
                alert('Ошибка при удалении комментария');
            }
        } catch (error) {
            console.error('Error deleting comment:', error);
            alert('Ошибка при удалении комментария: ' + error.message);
        }
    }

    async searchComments() {
        const query = document.getElementById('search').value.trim();
        if (!query) {
            this.loadComments();
            return;
        }

        try {
            const response = await fetch(`${this.baseUrl}/comments?search=${encodeURIComponent(query)}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            this.renderComments(data.comments);
        } catch (error) {
            console.error('Error searching comments:', error);
            alert('Ошибка поиска: ' + error.message);
        }
    }

    clearSearch() {
        document.getElementById('search').value = '';
        this.loadComments();
    }

    escapeHtml(unsafe) {
        if (!unsafe) return '';
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
}

const app = new CommentApp();

window.searchComments = () => app.searchComments();
window.clearSearch = () => app.clearSearch();