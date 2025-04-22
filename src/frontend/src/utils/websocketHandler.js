import { useWebSocket } from "@vueuse/core";
import { getToken } from "./tokenHandler";
import { useBulkStore } from "src/stores/bulkStore";
import { useWebStore } from "src/stores/webStore";
import { Notify } from "quasar";

var wsUrl = "/app/ws";
if (process.env.DEV) {
    wsUrl = process.env.QENV_API_URL + "/app/ws";
}

const token = getToken();
if (token) {
    wsUrl = wsUrl + "?token=" + token;
}

const bulkStore = useBulkStore();
const webStore = useWebStore();

export const { status, send, open, close } = useWebSocket(wsUrl, {
    autoReconnect: true,
    heartbeat: {
        message: JSON.stringify({ event: "ping" }),
        responseMessage: JSON.stringify({ event: "pong" }),
        interval: 60000, //60秒
        pongTimeout: 5000 //5秒
    },
    onConnected: (ws) => {
        console.log("Connected to websocket server: ", ws.url);
    },
    onDisconnected: (ws, event) => {
        console.log("Disconnected from websocket server with code ", event.code, " reason: ", event.reason);
    },
    onError: (ws, event) => {
        console.error("Got error on websocket eventPhase: ", event.eventPhase);
    },
    onMessage: (ws, msg) => {
        const msgObj = JSON.parse(msg.data);
        switch (msgObj.event) {
            case "bulkCheckInfo":
                bulkStore.setBulkCheckStatus(msgObj.data.Status);
                bulkStore.setBulkCheckQueryType(msgObj.data.QueryType);
                bulkStore.updateBulkCheckInfo(msgObj.data);
                break;
            case "webCheckDomains":
                webStore.setCheckingDomains(msgObj.data, "webCheck", "");
                break;
            case "typoResult":
                webStore.setCheckingDomains(msgObj.data.domains, "typoCheck", msgObj.data.typoType);
                break;
            case "webCheckResult":
                webStore.updateCheckResult(msgObj.data);
                break;
            case "registerResult":
                webStore.updateRegisterResult(msgObj.data);
                break;
            case "bulkCheckError":
            case "webCheckError":
            case "typoCheckError":
            case "registerError":
                Notify.create({
                    position: "top",
                    type: "negative",
                    message: "错误: " + msgObj.data,
                    progress: true,
                    timeout: 10000,
                    actions: [{ label: "确定", color: "white" }]
                });
                break;
        }
    }
});
