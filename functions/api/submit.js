export async function onRequestPost(context) {
    const { request, env } = context;
    const formData = await request.formData();
    const message = formData.get("name");
    if (!message) {
        return new Response("Missing message", { status: 400 });
    }
    // Store in KV with key "msg"
    await env.MESSAGES.put("msg", message);
    return new Response("Message stored!", { status: 200 });
}
