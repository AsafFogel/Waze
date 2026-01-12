import xml.etree.ElementTree as ET
import json
import math
from collections import deque, defaultdict

INPUT_FILE = '../data/export.osm'
OUTPUT_FILE = '../data/new_shoham.json'
DEFAULT_SPEED = 50.0

def haversine_distance(lat1, lon1, lat2, lon2):
    R = 6371.0 
    dlat = math.radians(lat2 - lat1)
    dlon = math.radians(lon2 - lon1)
    a = math.sin(dlat / 2)**2 + math.cos(math.radians(lat1)) * math.cos(math.radians(lat2)) * math.sin(dlon / 2)**2
    c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))
    return R * c

def keep_only_largest_component(nodes_list, edges_list):
    adj = defaultdict(list)
    for e in edges_list:
        u, v = e['from'], e['to']
        adj[u].append(v)
        adj[v].append(u)

    visited = set()
    components = [] 

    for node in nodes_list:
        node_id = node['id']
        if node_id not in visited:

            component = []
            queue = deque([node_id])
            visited.add(node_id)
            while queue:
                curr = queue.popleft()
                component.append(curr)
                for neighbor in adj[curr]:
                    if neighbor not in visited:
                        visited.add(neighbor)
                        queue.append(neighbor)
            components.append(component)

    if not components:
        return [], []

    largest_component = max(components, key=len)
    print(f"Graph check: Found {len(components)} disconnected islands.")
    print(f"Keeping largest component with {len(largest_component)} nodes (removing {len(nodes_list) - len(largest_component)} nodes).")

    valid_node_ids = set(largest_component)
    
    new_nodes = [n for n in nodes_list if n['id'] in valid_node_ids]
    new_edges = [e for e in edges_list if e['from'] in valid_node_ids and e['to'] in valid_node_ids]

    return new_nodes, new_edges

# --------------------------------------------

print(f"Parsing OSM file: {INPUT_FILE}...")
try:
    tree = ET.parse(INPUT_FILE)
    root = tree.getroot()
except FileNotFoundError:
    print("Error: File not found.")
    exit()

nodes = {}
internal_id_map = {} 
next_node_id = 1

print("Loading nodes...")
for node in root.findall('node'):
    osm_id = node.get('id')
    lat = float(node.get('lat'))
    lon = float(node.get('lon'))
    nodes[osm_id] = {'x': lon, 'y': lat} 

print("Processing edges...")
final_nodes_map = {}
final_edges = []
edge_id_counter = 1

for way in root.findall('way'):
    highway = way.find("tag[@k='highway']")
    if highway is None:
        continue
    
    road_type = highway.get('v')
    
    if road_type in ['footway', 'cycleway', 'path', 'steps', 'pedestrian', 'track', 'service']:
        continue

    nd_refs = [nd.get('ref') for nd in way.findall('nd')]
    
    oneway = False
    oneway_tag = way.find("tag[@k='oneway']")
    if oneway_tag is not None and oneway_tag.get('v') == 'yes':
        oneway = True

    speed_limit = 50.0
    maxspeed = way.find("tag[@k='maxspeed']")
    if maxspeed is not None:
        try:
            speed_limit = float(maxspeed.get('v').split()[0])
        except:
            pass

    for i in range(len(nd_refs) - 1):
        u_osm = nd_refs[i]
        v_osm = nd_refs[i+1]
        
        if u_osm not in nodes or v_osm not in nodes:
            continue

        # Remap IDs
        if u_osm not in internal_id_map:
            internal_id_map[u_osm] = next_node_id
            final_nodes_map[next_node_id] = {'id': next_node_id, **nodes[u_osm]}
            next_node_id += 1
        
        if v_osm not in internal_id_map:
            internal_id_map[v_osm] = next_node_id
            final_nodes_map[next_node_id] = {'id': next_node_id, **nodes[v_osm]}
            next_node_id += 1
            
        u_id = internal_id_map[u_osm]
        v_id = internal_id_map[v_osm]

        dist = haversine_distance(nodes[u_osm]['y'], nodes[u_osm]['x'], 
                                  nodes[v_osm]['y'], nodes[v_osm]['x'])
        
        if dist < 0.001: dist = 0.001

        # Add Edge
        final_edges.append({
            "id": edge_id_counter,
            "from": u_id,
            "to": v_id,
            "length": dist,
            "speedLimit": speed_limit
        })
        edge_id_counter += 1

        if not oneway:
            final_edges.append({
                "id": edge_id_counter,
                "from": v_id,
                "to": u_id,
                "length": dist,
                "speedLimit": speed_limit
            })
            edge_id_counter += 1


nodes_list = list(final_nodes_map.values())


print("Cleaning disconnected islands...")
clean_nodes, clean_edges = keep_only_largest_component(nodes_list, final_edges)

output = {"nodes": clean_nodes, "edges": clean_edges}

with open(OUTPUT_FILE, "w") as f:
    json.dump(output, f, indent=2)

print(f"Success! Created {OUTPUT_FILE}")
print(f"   Final Nodes: {len(clean_nodes)}")
print(f"   Final Edges: {len(clean_edges)}")