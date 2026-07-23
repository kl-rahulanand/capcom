from typing_extensions import TypedDict

from langgraph.graph import END, START, StateGraph


class DemoState(TypedDict, total=False):
    message: str
    response: str


def respond(state: DemoState) -> DemoState:
    message = state.get("message", "no message supplied")
    return {"response": f"Capcom observed LangGraph: {message}"}


builder = StateGraph(DemoState)
builder.add_node("respond", respond)
builder.add_edge(START, "respond")
builder.add_edge("respond", END)

graph = builder.compile()
