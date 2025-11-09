const API = 'http://localhost:8080';

async function uploadFile(ev) {
  ev.preventDefault();
  const f = document.getElementById('fileInput').files[0];
  if (!f) { alert('choose file'); return; }
  const fd = new FormData();
  fd.append('file', f);
  const res = await fetch(API + '/upload', { method: 'POST', body: fd });
  if (!res.ok) {
    alert('upload failed: ' + res.status);
    return;
  }
  const j = await res.json();
  addLocalId(j.id);
  refresh();
}

function addLocalId(id) {
  const arr = JSON.parse(localStorage.getItem('images') || '[]');
  arr.unshift(id);
  localStorage.setItem('images', JSON.stringify(arr));
}

async function refresh() {
  const list = JSON.parse(localStorage.getItem('images') || '[]');
  const cont = document.getElementById('list');
  cont.innerHTML = '';
  for (const id of list) {
    const card = document.createElement('div');
    card.className = 'card';
    const img = document.createElement('img');
    img.src = 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300"><rect width="100%" height="100%" fill="%23eee"/><text x="20" y="40">In processing...</text></svg>';
    const meta = document.createElement('div');
    meta.className = 'meta';
    const span = document.createElement('span');
    span.textContent = id;
    const btn = document.createElement('button');
    btn.textContent = 'Details';
    btn.onclick = async () => {
      const r = await fetch(API + '/image/' + id);
      if (r.status !== 200) return alert('not found');
      const j = await r.json();
      if (j.status === 'done') {
        const v = j.versions;
        let url = null;
        for (const fname in v) {
          if (fname.includes('watermark') || fname.includes('watermarked')) {
            url = v[fname];
            break;
          }
        }
        if (!url) {
          for (const fname in v) {
            if (fname.includes('resized')) { url = v[fname]; break; }
          }
        }
        if (!url && j.original) url = j.original;
        if (url) img.src = url;
      } else {
        alert('status: ' + j.status);
      }
    };
    const del = document.createElement('button');
    del.textContent = 'Delete';
    del.onclick = async () => {
      const r = await fetch(API + '/image/' + id, { method: 'DELETE' });
      if (r.status === 200) {
        const arr = JSON.parse(localStorage.getItem('images') || '[]').filter(x => x !== id);
        localStorage.setItem('images', JSON.stringify(arr));
        refresh();
      } else {
        alert('delete failed');
      }
    };
    meta.appendChild(span);
    meta.appendChild(btn);
    meta.appendChild(del);
    card.appendChild(img);
    card.appendChild(meta);
    cont.appendChild(card);
  }
}

document.getElementById('uploadForm').addEventListener('submit', uploadFile);
document.getElementById('refresh').addEventListener('click', refresh);
refresh();