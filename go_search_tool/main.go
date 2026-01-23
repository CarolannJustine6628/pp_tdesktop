package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var indexer *Indexer

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Telegram Local Search</title>
    <meta charset="UTF-8">
    <style>
        * { box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; 
            margin: 0; 
            padding: 0; 
            height: 100vh; 
            display: flex; 
            flex-direction: column; 
            background: #f0f2f5;
        }
        header { 
            padding: 20px; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-bottom: 1px solid #ddd;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .search-box { 
            display: flex; 
            gap: 12px; 
            max-width: 900px; 
            margin: 0 auto; 
        }
        input { 
            flex-grow: 1; 
            padding: 14px 18px; 
            font-size: 16px; 
            border: none;
            border-radius: 12px;
            background: white;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            transition: box-shadow 0.3s;
        }
        input:focus {
            outline: none;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        button { 
            padding: 14px 28px; 
            font-size: 16px; 
            cursor: pointer;
            background: white;
            border: none;
            border-radius: 12px;
            font-weight: 600;
            color: #667eea;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            transition: all 0.3s;
        }
        button:hover {
            background: #f8f9fa;
            transform: translateY(-1px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.15);
        }
        
        #container { flex: 1; display: flex; overflow: hidden; }
        
        #sidebar { 
            width: 320px; 
            border-right: 1px solid #e0e0e0; 
            overflow-y: auto; 
            background: white;
            box-shadow: 2px 0 8px rgba(0,0,0,0.05);
        }
        .peer-item {
            padding: 18px 20px;
            border-bottom: 1px solid #f0f0f0;
            cursor: pointer;
            transition: all 0.2s;
            position: relative;
        }
        .peer-item:hover { 
            background: #f8f9fa; 
            padding-left: 24px;
        }
        .peer-item.active { 
            background: linear-gradient(90deg, #667eea15 0%, transparent 100%);
            border-left: 4px solid #667eea;
            font-weight: 600;
        }
        .peer-id { 
            font-size: 0.95em; 
            color: #333; 
            display: block;
            margin-bottom: 4px;
        }
        .peer-count { 
            font-size: 0.85em; 
            color: #667eea; 
            font-weight: 600;
            background: #f0f2f5;
            padding: 4px 10px;
            border-radius: 12px;
            display: inline-block;
        }

        #main { 
            flex: 1; 
            overflow-y: auto; 
            padding: 24px; 
            background: #f0f2f5;
        }
        
        .result { 
            background: white;
            border: none;
            padding: 20px;
            margin-bottom: 16px;
            border-radius: 16px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.08);
            transition: all 0.3s;
            cursor: pointer;
            position: relative;
        }
        .result:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 16px rgba(0,0,0,0.12);
        }
        .result.expanded {
            box-shadow: 0 8px 24px rgba(0,0,0,0.15);
        }
        .meta { 
            color: #666; 
            font-size: 0.9em; 
            margin-bottom: 12px; 
            padding-bottom: 12px; 
            border-bottom: 1px solid #f0f0f0;
            display: flex;
            align-items: center;
            gap: 12px;
            flex-wrap: wrap;
        }
        .meta-item {
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .meta-label {
            color: #999;
            font-size: 0.85em;
        }
        .meta-value {
            color: #667eea;
            font-weight: 600;
        }
        .text { 
            white-space: pre-wrap; 
            font-size: 1.05em; 
            line-height: 1.7; 
            color: #333;
            word-break: break-word;
        }
        
        .context-section {
            margin-top: 20px;
            padding-top: 20px;
            border-top: 2px dashed #e0e0e0;
            animation: fadeIn 0.4s ease-in;
        }
        .context-header {
            font-size: 0.9em;
            color: #667eea;
            font-weight: 600;
            margin-bottom: 16px;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .context-messages {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }
        .context-msg {
            background: #f8f9fa;
            padding: 14px 18px;
            border-radius: 12px;
            border-left: 3px solid #e0e0e0;
            transition: all 0.2s;
        }
        .context-msg:hover {
            background: #f0f2f5;
            border-left-color: #667eea;
        }
        .context-msg.target {
            background: linear-gradient(90deg, #667eea15 0%, transparent 100%);
            border-left-color: #667eea;
            border-left-width: 4px;
            font-weight: 500;
        }
        .context-msg-meta {
            font-size: 0.85em;
            color: #888;
            margin-bottom: 6px;
            display: flex;
            gap: 12px;
            align-items: center;
        }
        .context-msg-text {
            color: #444;
            line-height: 1.6;
            white-space: pre-wrap;
            word-break: break-word;
        }
        .loading {
            text-align: center;
            padding: 20px;
            color: #667eea;
        }
        .loading::after {
            content: '...';
            animation: dots 1.5s steps(4, end) infinite;
        }
        
        .empty-state { 
            text-align: center; 
            color: #999; 
            margin-top: 80px;
            font-size: 1.1em;
        }
        .empty-state-icon {
            font-size: 4em;
            margin-bottom: 16px;
            opacity: 0.5;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        @keyframes dots {
            0%, 20% { content: '.'; }
            40% { content: '..'; }
            60%, 100% { content: '...'; }
        }
        
        .close-context {
            position: absolute;
            top: 16px;
            right: 16px;
            background: #f0f0f0;
            border: none;
            border-radius: 50%;
            width: 32px;
            height: 32px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 18px;
            color: #666;
            transition: all 0.2s;
        }
        .close-context:hover {
            background: #e0e0e0;
            transform: rotate(90deg);
        }
    </style>
</head>
<body>
    <header>
        <div class="search-box">
            <input type="text" id="query" placeholder="搜索消息内容..." onkeypress="handleEnter(event)">
            <button onclick="doSearch()">🔍 搜索</button>
        </div>
    </header>
    <div id="container">
        <div id="sidebar"></div>
        <div id="main">
            <div class="empty-state">
                <div class="empty-state-icon">💬</div>
                <div>输入关键词开始搜索</div>
            </div>
        </div>
    </div>

    <script>
        let currentData = {};
        let expandedResult = null;

        function handleEnter(e) {
            if (e.key === 'Enter') doSearch();
        }

        async function doSearch() {
            const query = document.getElementById('query').value;
            if (!query.trim()) return;

            document.getElementById('sidebar').innerHTML = '<div class="loading">搜索中</div>';
            document.getElementById('main').innerHTML = '';
            expandedResult = null;

            try {
                const res = await fetch('/api/search?q=' + encodeURIComponent(query));
                currentData = await res.json();
                renderSidebar();
            } catch (error) {
                document.getElementById('main').innerHTML = '<div class="empty-state">搜索失败，请重试</div>';
            }
        }

        function renderSidebar() {
            const sidebar = document.getElementById('sidebar');
            sidebar.innerHTML = '';
            
            const peerIds = Object.keys(currentData).sort((a, b) => currentData[b].length - currentData[a].length);
            
            if (peerIds.length === 0) {
                sidebar.innerHTML = '<div style="padding:20px; color:#999;">未找到结果</div>';
                document.getElementById('main').innerHTML = '<div class="empty-state"><div class="empty-state-icon">🔍</div><div>未找到匹配的消息</div></div>';
                return;
            }

            peerIds.forEach(peerId => {
                const msgs = currentData[peerId];
                const div = document.createElement('div');
                div.className = 'peer-item';
                div.onclick = () => selectPeer(peerId, div);
                div.innerHTML = '<span class="peer-id">聊天 ' + peerId + '</span>' + 
                              '<span class="peer-count">' + msgs.length + ' 条</span>';
                sidebar.appendChild(div);
            });

            if (peerIds.length > 0) {
                sidebar.firstChild.click();
            }
        }

        function selectPeer(peerId, el) {
            document.querySelectorAll('.peer-item').forEach(e => e.classList.remove('active'));
            el.classList.add('active');

            const main = document.getElementById('main');
            main.innerHTML = '';
            expandedResult = null;
            
            const msgs = currentData[peerId];
            main.scrollTop = 0;

            msgs.forEach(msg => {
                const div = document.createElement('div');
                div.className = 'result';
                div.dataset.peerId = peerId;
                div.dataset.msgId = msg.msg_id;
                
                const date = new Date(msg.date * 1000).toLocaleString('zh-CN');
                
                div.innerHTML = '<div class="meta">' +
                              '<div class="meta-item"><span class="meta-label">消息ID:</span><span class="meta-value">' + msg.msg_id + '</span></div>' +
                              '<div class="meta-item"><span class="meta-label">发送者:</span><span class="meta-value">' + msg.from_id + '</span></div>' +
                              '<div class="meta-item"><span class="meta-label">时间:</span><span class="meta-value">' + date + '</span></div>' +
                              '</div>' +
                              '<div class="text"></div>';
                div.querySelector('.text').textContent = msg.text;
                
                div.onclick = (e) => {
                    if (e.target.classList.contains('close-context')) return;
                    if (expandedResult === div) {
                        collapseContext(div);
                    } else {
                        if (expandedResult) collapseContext(expandedResult);
                        expandContext(div, peerId, msg.msg_id);
                    }
                };
                
                main.appendChild(div);
            });
        }

        async function expandContext(resultDiv, peerId, msgId) {
            if (expandedResult === resultDiv) return;
            
            expandedResult = resultDiv;
            resultDiv.classList.add('expanded');
            
            const contextDiv = document.createElement('div');
            contextDiv.className = 'context-section';
            contextDiv.innerHTML = '<div class="loading">加载上下文</div>';
            resultDiv.appendChild(contextDiv);

            try {
                const res = await fetch('/api/context?peer_id=' + peerId + '&msg_id=' + msgId + '&before=5&after=5');
                const contextMsgs = await res.json();

                let contextHTML = '<div class="context-header">📋 上下文消息 (' + contextMsgs.length + ' 条)</div>';
                contextHTML += '<div class="context-messages">';

                contextMsgs.forEach(msg => {
                    const isTarget = msg.msg_id == msgId;
                    const date = new Date(msg.date * 1000).toLocaleString('zh-CN');
                    contextHTML += '<div class="context-msg' + (isTarget ? ' target' : '') + '">';
                    contextHTML += '<div class="context-msg-meta">';
                    contextHTML += '<span>ID: ' + msg.msg_id + '</span>';
                    contextHTML += '<span>From: ' + msg.from_id + '</span>';
                    contextHTML += '<span>' + date + '</span>';
                    if (isTarget) contextHTML += '<span style="color:#667eea; font-weight:600;">← 当前消息</span>';
                    contextHTML += '</div>';
                    contextHTML += '<div class="context-msg-text">' + escapeHtml(msg.text || '(空消息)') + '</div>';
                    contextHTML += '</div>';
                });

                contextHTML += '</div>';
                contextHTML += '<button class="close-context" onclick="collapseContext(this.closest(\'.result\'))">×</button>';
                
                contextDiv.innerHTML = contextHTML;
            } catch (error) {
                contextDiv.innerHTML = '<div style="color:#f44336; padding:10px;">加载上下文失败</div>';
            }
        }

        function collapseContext(resultDiv) {
            if (!resultDiv) return;
            resultDiv.classList.remove('expanded');
            const contextSection = resultDiv.querySelector('.context-section');
            if (contextSection) {
                contextSection.remove();
            }
            if (expandedResult === resultDiv) {
                expandedResult = null;
            }
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
    </script>
</body>
</html>
`

func main() {
	dataPath := flag.String("path", "tdata/local_message_index", "Path to index directory")
	port := flag.String("port", "8800", "HTTP Port")
	flag.Parse()

	indexer = NewIndexer()
	
	fmt.Printf("Loading index from %s...\n", *dataPath)
	if err := indexer.LoadAll(*dataPath); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}
	
	// Watch for changes (simple poll every 30s)
	go func() {
		for {
			time.Sleep(30 * time.Second)
			indexer.LoadAll(*dataPath)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.New("index").Parse(htmlTemplate)
		t.Execute(w, nil)
	})

	http.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		results := indexer.Search(query)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	http.HandleFunc("/api/context", func(w http.ResponseWriter, r *http.Request) {
		var peerID uint64
		var msgID int64
		var beforeCount, afterCount int = 5, 5

		fmt.Sscanf(r.URL.Query().Get("peer_id"), "%d", &peerID)
		fmt.Sscanf(r.URL.Query().Get("msg_id"), "%d", &msgID)
		fmt.Sscanf(r.URL.Query().Get("before"), "%d", &beforeCount)
		fmt.Sscanf(r.URL.Query().Get("after"), "%d", &afterCount)

		if peerID == 0 || msgID == 0 {
			http.Error(w, "peer_id and msg_id are required", http.StatusBadRequest)
			return
		}

		context := indexer.GetContext(peerID, msgID, beforeCount, afterCount)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(context)
	})

	fmt.Printf("Server starting at http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
