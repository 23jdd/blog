new Vue({
  el: "#app",
  data: function () {
    return {
      tab: "dashboard",
      baseURL: localStorage.getItem("apiBaseURL") || "http://localhost:8080",
      token: localStorage.getItem("accessToken") || "",
      refreshToken: localStorage.getItem("refreshToken") || "",
      auth: { username: "alice", password: "P@ssw0rd123" },
      articleForm: { title: "", content_url: "", draft_id: 0, status: "published", category_id: 1, tags: "", cover_url: "" },
      draftForm: { title: "", content_url: "", content: "" },
      selectedArticleId: 0,
      selectedDraftId: 0,
      interaction: { article_id: 0, comment: "", parent_id: 0, comment_id: 0, comment_status: "approved" },
      cfg: { key: "limit.read", value: "200" },
      userForm: { age: "20", gender: "male" },
      userInfoPreview: null,
      myArticles: [],
      drafts: [],
      hotArticles: [],
      reader: { articleId: 0, title: "", contentURL: "", markdown: "", html: "" },
      draftEditor: { id: 0, title: "", contentURL: "", markdown: "", html: "" },
      markdownPreview: "",
      lastResponse: { message: "等待请求..." }
    };
  },
  mounted: function () {
    this.loadHot();
    this.loadDrafts();
    this.loadMyArticles();
  },
  methods: {
    pretty: function (v) { return JSON.stringify(v, null, 2); },
    saveBaseURL: function () {
      localStorage.setItem("apiBaseURL", this.baseURL);
      this.ok("已保存 API 地址");
    },
    setTokens: function (data) {
      this.token = data.access_token || this.token || "";
      this.refreshToken = data.refresh_token || this.refreshToken || "";
      localStorage.setItem("accessToken", this.token || "");
      localStorage.setItem("refreshToken", this.refreshToken || "");
    },
    clearTokens: function () {
      this.token = "";
      this.refreshToken = "";
      localStorage.removeItem("accessToken");
      localStorage.removeItem("refreshToken");
    },
    api: function () {
      var headers = {};
      if (this.token) headers.Authorization = "Bearer " + this.token;
      return axios.create({ baseURL: this.baseURL, headers: headers, timeout: 20000 });
    },
    ok: function (msg, data) { this.lastResponse = { code: 200, message: msg, data: data || null }; },
    resolveURL: function (u) {
      if (!u) return "";
      if (u.indexOf("http://") === 0 || u.indexOf("https://") === 0) return u;
      if (u.indexOf("/") === 0) return this.baseURL.replace(/\/$/, "") + u;
      return this.baseURL.replace(/\/$/, "") + "/" + u;
    },
    fetchTextByURL: async function (u) {
      var full = this.resolveURL(u);
      var res = await axios.get(full, { responseType: "text", timeout: 20000 });
      return typeof res.data === "string" ? res.data : JSON.stringify(res.data);
    },
    renderMarkdownHTML: async function (md) {
      var res = await this.api().post("/markdown", md, { headers: { "Content-Type": "text/plain" } });
      return typeof res.data === "string" ? res.data : (res.data && res.data.html) || "";
    },
    escapeHTML: function (s) {
      return String(s || "")
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");
    },
    err: function (e) {
      if (e && e.response) this.lastResponse = e.response.data || { status: e.response.status };
      else this.lastResponse = { message: String(e) };
    },

    register: async function () {
      try {
        var res = await this.api().post("/auth/register", this.auth);
        this.setTokens(res.data || {});
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    login: async function () {
      try {
        var res = await this.api().post("/auth/login", this.auth);
        this.setTokens(res.data || {});
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    refresh: async function () {
      try {
        var res = await this.api().post("/auth/refresh", { refresh_token: this.refreshToken });
        this.setTokens(res.data || {});
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    logout: async function () {
      try {
        var res = await this.api().post("/auth/logout", { refresh_token: this.refreshToken });
        this.clearTokens();
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },

    createArticle: async function () {
      try {
        var body = Object.assign({}, this.articleForm);
        if (!body.draft_id) delete body.draft_id;
        var res = await this.api().post("/articles", body);
        this.lastResponse = res.data;
        this.loadMyArticles();
      } catch (e) { this.err(e); }
    },
    updateArticle: async function () {
      try {
        if (!this.selectedArticleId) return this.ok("请先填写文章ID");
        var body = {
          title: this.articleForm.title,
          content_url: this.articleForm.content_url,
          category_id: this.articleForm.category_id,
          tags: this.articleForm.tags,
          cover_url: this.articleForm.cover_url
        };
        var res = await this.api().put("/articles/" + this.selectedArticleId, body);
        this.lastResponse = res.data;
        this.loadMyArticles();
      } catch (e) { this.err(e); }
    },
    deleteArticle: async function () {
      try {
        if (!this.selectedArticleId) return this.ok("请先填写文章ID");
        var res = await this.api().delete("/articles/" + this.selectedArticleId);
        this.lastResponse = res.data;
        this.loadMyArticles();
      } catch (e) { this.err(e); }
    },
    getArticle: async function () {
      try {
        if (!this.selectedArticleId) return this.ok("请先填写文章ID");
        var res = await this.api().get("/articles/" + this.selectedArticleId);
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    quickStatus: async function (a, status) {
      try {
        var res = await this.api().patch("/articles/" + a.id + "/status", { status: status });
        this.lastResponse = res.data;
        this.loadMyArticles();
      } catch (e) { this.err(e); }
    },
    loadMyArticles: async function () {
      try {
        if (!this.token) return;
        var judge = await this.api().get("/auth/judgeToken");
        this.lastResponse = judge.data;
        var uid = judge.data && judge.data.user_id;
        if (!uid) return;
        var res = await this.api().get("/articles/author/" + uid);
        this.myArticles = Array.isArray(res.data) ? res.data : [];
      } catch (e) { this.err(e); }
    },
    loadHot: async function () {
      try {
        var res = await this.api().get("/articles/hot", { params: { limit: 10 } });
        this.hotArticles = Array.isArray(res.data) ? res.data : (res.data.data || []);
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    fillArticle: function (a) {
      this.selectedArticleId = a.id;
      this.articleForm.title = a.title || "";
      this.articleForm.content_url = a.content || "";
      this.articleForm.status = a.status || "draft";
      this.articleForm.category_id = a.category_id || 0;
      this.articleForm.tags = a.tags || "";
      this.articleForm.cover_url = a.cover_url || "";
      this.articleForm.draft_id = 0;
      this.tab = "articles";
    },
    openReaderFromArticle: async function (a) {
      this.reader.articleId = a.id || 0;
      this.reader.title = a.title || "";
      this.reader.contentURL = a.content || "";
      this.tab = "reader";
      await this.readByURL();
    },
    readArticleById: async function () {
      try {
        if (!this.reader.articleId) return this.ok("请先填写文章ID");
        var res = await this.api().get("/articles/" + this.reader.articleId);
        this.lastResponse = res.data;
        this.reader.title = res.data.title || "";
        this.reader.contentURL = res.data.content || "";
        await this.readByURL();
      } catch (e) { this.err(e); }
    },
    readByURL: async function () {
      try {
        if (!this.reader.contentURL) return this.ok("请先填写 content_url");
        this.reader.markdown = await this.fetchTextByURL(this.reader.contentURL);
        this.reader.html = await this.renderMarkdownHTML(this.reader.markdown);
        this.ok("文章读取成功", { content_url: this.reader.contentURL });
      } catch (e) { this.err(e); }
    },

    saveDraft: async function () {
      try {
        var body = {
          title: this.draftForm.title
        };
        if ((this.draftForm.content || "").trim()) body.content = this.draftForm.content;
        if ((this.draftForm.content_url || "").trim()) body.content_url = this.draftForm.content_url;
        var res = await this.api().post("/drafts", body);
        this.lastResponse = res.data;
        this.loadDrafts();
      } catch (e) { this.err(e); }
    },
    updateDraft: async function () {
      try {
        if (!this.selectedDraftId) return this.ok("请先填写草稿ID");
        var body = {
          title: this.draftForm.title
        };
        if ((this.draftForm.content || "").trim()) body.content = this.draftForm.content;
        if ((this.draftForm.content_url || "").trim()) body.content_url = this.draftForm.content_url;
        var res = await this.api().put("/drafts/" + this.selectedDraftId, body);
        this.lastResponse = res.data;
        this.loadDrafts();
      } catch (e) { this.err(e); }
    },
    deleteDraft: async function () {
      try {
        if (!this.selectedDraftId) return this.ok("请先填写草稿ID");
        var res = await this.api().delete("/drafts/" + this.selectedDraftId);
        this.lastResponse = res.data;
        this.loadDrafts();
      } catch (e) { this.err(e); }
    },
    loadDrafts: async function () {
      try {
        if (!this.token) return;
        var res = await this.api().get("/drafts");
        this.lastResponse = res.data;
        this.drafts = (res.data && res.data.data && res.data.data.list) || [];
      } catch (e) { this.err(e); }
    },
    fillDraft: function (d) {
      this.selectedDraftId = d.id;
      this.draftForm.title = d.title || "";
      this.draftForm.content_url = d.content || "";
      this.draftForm.content = "";
    },
    fillDraftContentFromURL: async function (d) {
      try {
        this.selectedDraftId = d.id;
        this.draftForm.title = d.title || "";
        this.draftForm.content_url = d.content || "";
        this.draftForm.content = await this.fetchTextByURL(d.content || "");
        this.ok("已从草稿 URL 加载内容到文本框");
      } catch (e) { this.err(e); }
    },
    openEditor: async function (d) {
      this.draftEditor.id = d.id;
      this.draftEditor.title = d.title || "";
      this.draftEditor.contentURL = d.content || "";
      this.tab = "editor";
      await this.loadEditorByURL();
    },
    loadDraftToEditor: async function () {
      try {
        if (!this.draftEditor.id) return this.ok("请先填写草稿ID");
        var res = await this.api().get("/drafts");
        var list = (res.data && res.data.data && res.data.data.list) || [];
        var d = list.find(function (x) { return x.id === Number(this.draftEditor.id); }.bind(this));
        if (!d) return this.ok("未找到草稿，请先刷新列表");
        this.draftEditor.title = d.title || "";
        this.draftEditor.contentURL = d.content || "";
        await this.loadEditorByURL();
      } catch (e) { this.err(e); }
    },
    loadEditorByURL: async function () {
      try {
        if (!this.draftEditor.contentURL) return this.ok("请先填写草稿 content_url");
        this.draftEditor.markdown = await this.fetchTextByURL(this.draftEditor.contentURL);
        this.draftEditor.html = await this.renderMarkdownHTML(this.draftEditor.markdown);
        this.ok("草稿内容加载成功", { content_url: this.draftEditor.contentURL });
      } catch (e) { this.err(e); }
    },
    previewEditorMarkdown: async function () {
      this.draftEditor.html = await this.renderMarkdownHTML(this.draftEditor.markdown || "");
      this.ok("草稿预览完成");
    },
    saveEditorAsNewFile: async function () {
      try {
        if (!this.draftEditor.id) return this.ok("请先填写草稿ID");
        if (!this.draftEditor.markdown.trim()) return this.ok("编辑器内容不能为空");
        var upload = await this.api().post("/file/uploadArticle", this.draftEditor.markdown, { headers: { "Content-Type": "text/plain", Authorization: "Bearer " + this.token } });
        var newURL = upload.data && upload.data.data && upload.data.data.url;
        if (!newURL) return this.ok("上传成功但未返回 URL");
        this.draftEditor.contentURL = newURL;
        var body = { title: this.draftEditor.title || "未命名草稿", content_url: newURL };
        var update = await this.api().put("/drafts/" + this.draftEditor.id, body);
        this.lastResponse = { upload: upload.data, update: update.data };
        await this.loadDrafts();
        await this.previewEditorMarkdown();
      } catch (e) { this.err(e); }
    },
    publish: async function (d) {
      try {
        var body = {
          draft_id: d.id,
          status: "published",
          title: d.title,
          content_url: d.content,
          category_id: 1,
          tags: "",
          cover_url: ""
        };
        var res = await this.api().post("/articles", body);
        this.lastResponse = res.data;
        this.loadMyArticles();
      } catch (e) { this.err(e); }
    },
    publishByDraft: async function () {
      if (!this.selectedDraftId) return this.ok("请先填写草稿ID");
      var d = this.drafts.find(function (x) { return x.id === Number(this.selectedDraftId); }.bind(this));
      if (!d) d = { id: this.selectedDraftId, title: this.draftForm.title, content: this.draftForm.content_url };
      await this.publish(d);
    },

    uploadAvatar: async function () {
      try {
        var f = this.$refs.avatarInput.files[0];
        if (!f) return this.ok("请选择头像文件");
        var fd = new FormData();
        fd.append("image", f);
        var res = await this.api().post("/file/setPersonImage", fd, { headers: { "Content-Type": "multipart/form-data", Authorization: "Bearer " + this.token } });
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    uploadMarkdown: async function () {
      try {
        var f = this.$refs.mdInput.files[0];
        if (!f) return this.ok("请选择 markdown 文件");
        var txt = await f.text();
        var res = await this.api().post("/file/uploadArticle", txt, { headers: { "Content-Type": "text/plain", Authorization: "Bearer " + this.token } });
        this.lastResponse = res.data;
        if (res.data && res.data.data && res.data.data.url) {
          this.articleForm.content_url = res.data.data.url;
          this.draftForm.content_url = res.data.data.url;
          this.draftForm.content = txt;
        }
      } catch (e) { this.err(e); }
    },

    createComment: async function () {
      try {
        if (!this.interaction.article_id) return this.ok("请填写文章ID");
        var body = { content: this.interaction.comment, parent_id: this.interaction.parent_id || 0 };
        var res = await this.api().post("/articles/" + this.interaction.article_id + "/comments", body);
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    listComments: async function () {
      try {
        if (!this.interaction.article_id) return this.ok("请填写文章ID");
        var res = await this.api().get("/articles/" + this.interaction.article_id + "/comments", { params: { page: 1, pageSize: 20 } });
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    auditComment: async function () {
      try {
        if (!this.interaction.comment_id) return this.ok("请填写评论ID");
        var res = await this.api().patch("/articles/comments/" + this.interaction.comment_id + "/status", { status: this.interaction.comment_status });
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    deleteComment: async function () {
      try {
        if (!this.interaction.comment_id) return this.ok("请填写评论ID");
        var res = await this.api().delete("/articles/comments/" + this.interaction.comment_id);
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    like: async function () {
      try {
        if (!this.interaction.article_id) return this.ok("请填写文章ID");
        var res = await this.api().post("/articles/" + this.interaction.article_id + "/likes");
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    unlike: async function () {
      try {
        if (!this.interaction.article_id) return this.ok("请填写文章ID");
        var res = await this.api().delete("/articles/" + this.interaction.article_id + "/likes");
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    collect: async function () {
      try {
        if (!this.interaction.article_id) return this.ok("请填写文章ID");
        var res = await this.api().post("/articles/" + this.interaction.article_id + "/collections");
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    uncollect: async function () {
      try {
        if (!this.interaction.article_id) return this.ok("请填写文章ID");
        var res = await this.api().delete("/articles/" + this.interaction.article_id + "/collections");
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    loadCollections: async function () {
      try {
        var res = await this.api().get("/interactions/my-collections", { params: { page: 1, pageSize: 20 } });
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    loadFeed: async function () {
      try {
        var res = await this.api().get("/interactions/feed", { params: { page: 1, pageSize: 20 } });
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },

    setConfig: async function () {
      try {
        var res = await this.api().post("/config", this.cfg);
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    getConfig: async function () {
      try {
        var res = await this.api().get("/config", { params: { key: this.cfg.key } });
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    previewMarkdown: async function () {
      try {
        var md = "# 预览\\n\\n这是一段 **Markdown** 预览。\\n\\n- Vue2\\n- Blog\\n";
        var res = await this.api().post("/markdown", md, { headers: { "Content-Type": "text/plain" } });
        this.markdownPreview = res.data;
        this.ok("markdown 预览成功");
      } catch (e) { this.err(e); }
    },

    getUserInfo: async function () {
      try {
        var res = await this.api().get("/user/info");
        this.userInfoPreview = res.data;
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    updateUserInfo: async function () {
      try {
        var fd = new FormData();
        fd.append("age", this.userForm.age || "");
        fd.append("gender", this.userForm.gender || "");
        var res = await this.api().post("/user/update", fd);
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    deleteUser: async function () {
      try {
        var res = await this.api().delete("/user/delete");
        this.lastResponse = res.data;
      } catch (e) { this.err(e); }
    },
    runUserSmokeTest: async function () {
      try {
        var steps = [];
        var r1 = await this.api().get("/user/info");
        steps.push({ step: "GET /user/info", data: r1.data });
        var fd = new FormData();
        fd.append("age", this.userForm.age || "20");
        fd.append("gender", this.userForm.gender || "male");
        var r2 = await this.api().post("/user/update", fd);
        steps.push({ step: "POST /user/update", data: r2.data });
        var r3 = await this.api().get("/user/info");
        steps.push({ step: "GET /user/info(after update)", data: r3.data });
        this.userInfoPreview = r3.data;
        this.lastResponse = { message: "user smoke test done", steps: steps };
      } catch (e) { this.err(e); }
    }
  }
});
