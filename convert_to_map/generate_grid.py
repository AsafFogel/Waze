import json
import math

# הגדרות הגריד
WIDTH = 10   # רוחב הגריד (מספר צמתים)
HEIGHT = 10  # גובה הגריד
EDGE_SPEED = 50.0 # מהירות ברירת מחדל

nodes = []
edges = []
edge_id_counter = 0

# 1. יצירת צמתים (Nodes)
for y in range(HEIGHT):
    for x in range(WIDTH):
        node_id = y * WIDTH + x
        nodes.append({
            "id": node_id,
            # קואורדינטות פיקטיביות (רק בשביל ה-Heuristic)
            "x": y * 0.01, 
            "y": x * 0.01 
        })

# 2. יצירת קשתות (Edges) - דו כיווניות
for y in range(HEIGHT):
    for x in range(WIDTH):
        u = y * WIDTH + x
        
        # חיבור לימין (אם יש)
        if x < WIDTH - 1:
            v = y * WIDTH + (x + 1)
            # הלוך
            edges.append({
                "id": edge_id_counter,
                "from": u, "to": v,
                "length": 1.0, # 1 ק"מ
                "speed_limit": EDGE_SPEED
            })
            edge_id_counter += 1
            # חזור
            edges.append({
                "id": edge_id_counter,
                "from": v, "to": u,
                "length": 1.0,
                "speed_limit": EDGE_SPEED
            })
            edge_id_counter += 1

        # חיבור למטה (אם יש)
        if y < HEIGHT - 1:
            v = (y + 1) * WIDTH + x
            # הלוך
            edges.append({
                "id": edge_id_counter,
                "from": u, "to": v,
                "length": 1.0,
                "speed_limit": EDGE_SPEED
            })
            edge_id_counter += 1
            # חזור
            edges.append({
                "id": edge_id_counter,
                "from": v, "to": u,
                "length": 1.0,
                "speed_limit": EDGE_SPEED
            })
            edge_id_counter += 1

# 3. שמירה לקובץ
output = {"nodes": nodes, "edges": edges}

filename = "../data/grid_map.json"
with open(filename, "w") as f:
    json.dump(output, f, indent=2)

print(f"Created {filename} with {len(nodes)} nodes and {len(edges)} edges.")
print(f"IDs range: Nodes 0-{len(nodes)-1}, Edges 0-{len(edges)-1}")